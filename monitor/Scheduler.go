package monitor

import (
	"math/rand"
	"sync"
	"time"

	"github.com/abrander/agento/logger"
	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/timeseries"
	"github.com/abrander/agento/userdb"
	"gopkg.in/mgo.v2/bson"
)

// NewScheduler will instantiate a new scheduler.
func NewScheduler(store Store, subject userdb.Subject) *Scheduler {
	return &Scheduler{
		store:   store,
		subject: subject,
	}
}

// Loop will simply loop through all monitors and emit changes and execute jobs.
func (s *Scheduler) Loop(wg sync.WaitGroup, serv timeseries.Database) {
	_, err := s.store.GetHost(userdb.God, "000000000000000000000000")
	if err != nil {
		// If localhost/localtransport does not exist. Do something.
		p, found := plugins.GetTransports()["localtransport"]
		if !found {
			logger.Red("monitor", "localtransport plugin not found")
		}
		host := &Host{
			Id:          bson.ObjectIdHex("000000000000000000000000"),
			AccountId:   bson.ObjectIdHex("000000000000000000000000"),
			Name:        "localhost",
			TransportId: "localtransport",
			Transport:   p().(plugins.Transport),
		}
		err = s.store.AddHost(nil, host)
		if err != nil {
			logger.Red("monitor", "Error inserting: %s", err.Error())
		}
		logger.Yellow("monitor", "Added localhost transport with id %s", host.Id.Hex())
	}

	ticker := time.Tick(time.Millisecond * 100)

	inFlight := make(map[bson.ObjectId]bool)
	inFlightLock := sync.RWMutex{}
	for t := range ticker {
		var monitors []Monitor
		// FIXME: Add a function to get *ALL* monitors regardless of account.
		monitors, err := s.store.GetAllMonitors(userdb.God, "000000000000000000000000")
		if err != nil {
			logger.Red("monitor", "Error getting monitors from Mongo: %s", err.Error())
			continue
		}

		for _, mon := range monitors {
			age := t.Sub(mon.LastCheck)  // positive: past
			wait := mon.NextCheck.Sub(t) // positive: future

			inFlightLock.RLock()
			_, found := inFlight[mon.Id]
			inFlightLock.RUnlock()

			if found {
				// skipping monitors in flight
			} else if age > mon.Interval*2 && wait < -mon.Interval {
				checkIn := time.Duration(rand.Int63n(int64(mon.Interval)))
				mon.NextCheck = t.Add(checkIn)
				logger.Yellow("monitor", "%s %s: Delaying first check by %s", mon.Id.Hex(), mon.Job.AgentId, checkIn)

				err = s.store.UpdateMonitor(userdb.God, &mon)
				if err != nil {
					logger.Red("Error updating: %v", err.Error())
				}
			} else if wait < 0 {
				inFlightLock.Lock()
				inFlight[mon.Id] = true
				inFlightLock.Unlock()

				go func(mon Monitor) {
					host, err := s.store.GetHost(userdb.God, mon.HostId.Hex())
					if err != nil {
						logger.Yellow("monitor", "GetHost(): %s", err.Error())
					}

					p, err := mon.Job.Run(host.Transport.(plugins.Transport))
					if err == nil {
						logger.Green("monitor", "%s, %s", mon.Id.Hex(), mon.Job.AgentId)
					} else {
						logger.Red("monitor", "%s, %s: %s", mon.Id.Hex(), mon.Job.AgentId, err.Error())
					}
					mon.LastPoints = p
					mon.LastCheck = t
					mon.NextCheck = t.Add(mon.Interval)

					err = s.store.UpdateMonitor(userdb.God, &mon)
					if err != nil {
						logger.Red("monitor", "Error updating: %s", err.Error())
					}

					err = serv.WritePoints(p)
					if err != nil {
						logger.Red("monitor", "Influxdb error: %s", err.Error())
					}
					inFlightLock.Lock()
					delete(inFlight, mon.Id)
					inFlightLock.Unlock()
				}(mon)
			}
		}
	}

	wg.Done()
}
