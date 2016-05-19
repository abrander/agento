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

// NewScheduler will instantiate a new scheduler. The scheduler needs a Store to
// read/write checks. If the system is not a multiuser system, userdb.God can be
// used as subject.
func NewScheduler(store Store, subject userdb.Subject) *Scheduler {
	return &Scheduler{
		store:   store,
		subject: subject,
	}
}

// Loop will simply loop through all monitors and emit changes and execute jobs.
func (s *Scheduler) Loop(wg sync.WaitGroup, serv timeseries.Database) {
	// Make sure we have the magic localhost. Maybe we should move this somewhere else.
	_, err := s.store.GetHost(s.subject, "000000000000000000000000")
	if err != nil {
		// If localhost/localtransport does not exist. Do something.
		p, found := plugins.GetTransports()["localtransport"]
		if !found {
			logger.Red("monitor", "localtransport plugin not found")
		}

		// Construct the magic host.
		host := &Host{
			Id:          bson.ObjectIdHex("000000000000000000000000"),
			AccountId:   bson.ObjectIdHex("000000000000000000000000"),
			Name:        "localhost",
			TransportId: "localtransport",
			Transport:   p().(plugins.Transport),
		}

		// Save it.
		err = s.store.AddHost(nil, host)
		if err != nil {
			logger.Red("monitor", "Error inserting: %s", err.Error())
			wg.Done()
			return
		}

		logger.Yellow("monitor", "Added localhost transport with id %s", host.Id.Hex())
	}

	// We tick ten times a second, this should be enough for now.
	ticker := time.Tick(time.Millisecond * 100)

	// inFlight is a list of monitor id's currently running
	inFlight := make(map[bson.ObjectId]bool)
	inFlightLock := sync.RWMutex{}
	for t := range ticker {
		// We start by extracting a list of all monitors. If this gets too
		// expensive at some point, we can do it less frequent.
		var monitors []Monitor
		// FIXME: Add a function to get *ALL* monitors regardless of account.
		monitors, err := s.store.GetAllMonitors(s.subject, "000000000000000000000000")
		if err != nil {
			logger.Red("monitor", "Error getting monitors from store: %s", err.Error())
			continue
		}

		// We iterate the list of monitors, to see if anything needs to be done.
		for _, mon := range monitors {
			// Calculate the age of the last check, if the age is positive, it's
			// in the past.
			age := t.Sub(mon.LastCheck)

			// Calculate how much we should wait before executing the job. If
			// the value is positive, it's in the future.
			wait := mon.NextCheck.Sub(t)

			// Check if the monitor is already executing.
			inFlightLock.RLock()
			_, found := inFlight[mon.Id]
			inFlightLock.RUnlock()

			if found {
				continue
			}

			// If the check is older than two intervals, we treat it as new.
			if age > mon.Interval*2 && wait < -mon.Interval {
				checkIn := time.Duration(rand.Int63n(int64(mon.Interval)))
				mon.NextCheck = t.Add(checkIn)
				logger.Yellow("monitor", "%s %s: Delaying first check by %s", mon.Id.Hex(), mon.Job.AgentId, checkIn)

				err = s.store.UpdateMonitor(s.subject, &mon)
				if err != nil {
					logger.Red("Error updating: %v", err.Error())
				}
			} else if wait < 0 {
				// If we arrive here, wait is sub-zero, which means that we
				// should execute now.
				inFlightLock.Lock()
				inFlight[mon.Id] = true
				inFlightLock.Unlock()

				// Execute the monitor job in its own go routine.
				go func(mon Monitor) {
					host, err := s.store.GetHost(s.subject, mon.HostId.Hex())
					if err != nil {
						logger.Red("monitor", "GetHost(): %s", err.Error())
						return
					}

					// Run the job.
					p, err := mon.Job.Run(host.Transport.(plugins.Transport))
					if err == nil {
						// FIXME: Print some timing.
						logger.Green("monitor", "%s, %s", mon.Id.Hex(), mon.Job.AgentId)
					} else {
						// FIXME: Print some timing.
						logger.Red("monitor", "%s, %s: %s", mon.Id.Hex(), mon.Job.AgentId, err.Error())
					}

					// Save the result
					mon.LastPoints = p

					// Save the check time and schedule next check.
					mon.LastCheck = t
					mon.NextCheck = t.Add(mon.Interval)

					// Save everything back to store.
					err = s.store.UpdateMonitor(s.subject, &mon)
					if err != nil {
						logger.Red("monitor", "Error updating: %s", err.Error())
					}

					// Write results to TSDB.
					err = serv.WritePoints(p)
					if err != nil {
						logger.Red("monitor", "Influxdb error: %s", err.Error())
					}

					// Remove the monitor from inFlight map.
					inFlightLock.Lock()
					delete(inFlight, mon.Id)
					inFlightLock.Unlock()
				}(mon)
			}
		}
	}

	wg.Done()
}
