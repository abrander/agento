package monitor

import (
	"errors"
	"math/rand"
	"os"
	"sync"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/abrander/agento/configuration"
	"github.com/abrander/agento/logger"
	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/timeseries"
	"github.com/abrander/agento/userdb"
	"github.com/influxdata/influxdb/client/v2"
)

type (
	Admin interface {
		GetAllMonitors(subject userdb.Subject, accountId string) ([]Monitor, error)
		AddMonitor(subject userdb.Subject, mon *Monitor) error
		GetMonitor(subject userdb.Subject, id string) (*Monitor, error)
		UpdateMonitor(subject userdb.Subject, mon *Monitor) error
		DeleteMonitor(subject userdb.Subject, id string) error

		GetAllHosts(subject userdb.Subject, accountId string) ([]Host, error)
		AddHost(subject userdb.Subject, host *Host) error
		GetHost(subject userdb.Subject, id string) (*Host, error)
		GetHostByName(subject userdb.Subject, name string) (*Host, error)
		DeleteHost(subject userdb.Subject, id string) error
	}

	Monitor struct {
		Id         bson.ObjectId   `json:"id" bson:"_id"`
		AccountId  bson.ObjectId   `json:"accountId" bson:"accountId"`
		HostId     bson.ObjectId   `json:"hostId" bson:"hostId"`
		Interval   time.Duration   `json:"interval"`
		Job        Job             `json:"agent"` // FIXME: Rename json to "job" - maybe
		LastCheck  time.Time       `json:"lastCheck"`
		NextCheck  time.Time       `json:"nextCheck"`
		LastPoints []*client.Point `json:"lastResult"`
	}

	Scheduler struct {
		changes Broadcaster
	}
)

var (
	sess              *mgo.Session
	db                *mgo.Database
	hostCollection    *mgo.Collection
	monitorCollection *mgo.Collection

	ErrorInvalidId error = errors.New("Invalid id")
)

func Init(config configuration.MonitorConfiguration) {
	sess, err := mgo.Dial(config.Mongo.Url)
	if err != nil {
		logger.Error("monitor", "Can't connect to mongo, go error %v", err)
		os.Exit(1)
	}
	logger.Green("monitor", "Connected to mongo/%s at %s", config.Mongo.Database, config.Mongo.Url)

	db = sess.DB(config.Mongo.Database)
	hostCollection = db.C("hosts")
	monitorCollection = db.C("monitors")
}

func (m *Monitor) GetAccountId() string {
	return m.AccountId.Hex()
}

func NewScheduler(changes Broadcaster) *Scheduler {
	return &Scheduler{changes: changes}
}

func (s *Scheduler) GetAllMonitors(subject userdb.Subject, accountId string) ([]Monitor, error) {
	var monitors []Monitor

	err := subject.CanAccess(userdb.ObjectProxy(accountId))
	if err != nil {
		return []Monitor{}, err
	}

	err = monitorCollection.Find(bson.M{"accountId": bson.ObjectIdHex(accountId)}).All(&monitors)
	if err != nil {
		return []Monitor{}, err
	}

	return monitors, nil
}

func (s *Scheduler) GetMonitor(subject userdb.Subject, id string) (*Monitor, error) {
	var monitor Monitor

	if !bson.IsObjectIdHex(id) {
		return &monitor, ErrorInvalidId
	}

	err := monitorCollection.FindId(bson.ObjectIdHex(id)).One(&monitor)
	if err != nil {
		logger.Red("monitor", "Error getting monitor %s from Mongo: %s", id, err.Error())
		return &monitor, err
	}

	err = subject.CanAccess(&monitor)
	if err != nil {
		return nil, err
	}

	return &monitor, nil
}

func (s *Scheduler) UpdateMonitor(subject userdb.Subject, mon *Monitor) error {
	_, err := s.GetMonitor(subject, mon.Id.Hex())
	if err != nil {
		return err
	}

	s.changes.Broadcast("monchange", mon)

	return monitorCollection.UpdateId(mon.Id, mon)
}

func (s *Scheduler) AddMonitor(subject userdb.Subject, mon *Monitor) error {
	mon.Id = bson.NewObjectId()

	err := subject.CanAccess(mon)
	if err != nil {
		return err
	}

	s.changes.Broadcast("monadd", mon)

	return monitorCollection.Insert(mon)
}

func (s *Scheduler) DeleteMonitor(subject userdb.Subject, id string) error {
	if !bson.IsObjectIdHex(id) {
		return ErrorInvalidId
	}

	mon, err := s.GetMonitor(subject, id)
	if err != nil {
		return err
	}

	s.changes.Broadcast("mondelete", mon)

	return monitorCollection.RemoveId(bson.ObjectIdHex(id))
}

func (s *Scheduler) Loop(wg sync.WaitGroup, subject userdb.Subject, serv timeseries.Database) {
	_, err := s.GetHost(subject, "000000000000000000000000")
	if err != nil {
		p, found := plugins.GetTransports()["localtransport"]
		if !found {
			logger.Red("monitor", "localtransport plugin not found")
		}
		host := Host{
			Id:          bson.ObjectIdHex("000000000000000000000000"),
			AccountId:   bson.ObjectIdHex("000000000000000000000000"),
			Name:        "localhost",
			TransportId: "localtransport",
			Transport:   p().(plugins.Transport),
		}
		err = hostCollection.Insert(host)
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
		err := monitorCollection.Find(bson.M{}).All(&monitors)
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

				err = s.UpdateMonitor(subject, &mon)
				if err != nil {
					logger.Red("Error updating: %v", err.Error())
				}
			} else if wait < 0 {
				inFlightLock.Lock()
				inFlight[mon.Id] = true
				inFlightLock.Unlock()

				go func(mon Monitor) {
					var host Host
					hostCollection.FindId(mon.HostId).One(&host)
					p, err := mon.Job.Run(host.Transport)
					if err == nil {
						logger.Green("monitor", "%s, %s", mon.Id.Hex(), mon.Job.AgentId)
					} else {
						logger.Red("monitor", "%s, %s: %s", mon.Id.Hex(), mon.Job.AgentId, err.Error())
					}
					mon.LastPoints = p
					mon.LastCheck = t
					mon.NextCheck = t.Add(mon.Interval)

					err = s.UpdateMonitor(subject, &mon)
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
