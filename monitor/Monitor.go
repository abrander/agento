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
	"github.com/influxdb/influxdb/client"
)

type (
	Monitor struct {
		Id         bson.ObjectId  `json:"id" bson:"_id"`
		HostId     bson.ObjectId  `json:"hostId" bson:"hostId"`
		Interval   time.Duration  `json:"interval"`
		Job        Job            `json:"agent"` // FIXME: Rename json to "job" - maybe
		LastCheck  time.Time      `json:"lastCheck"`
		NextCheck  time.Time      `json:"nextCheck"`
		LastPoints []client.Point `json:"lastResult"`
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

	db = sess.DB(config.Mongo.Database)
	hostCollection = db.C("hosts")
	monitorCollection = db.C("monitors")
}

func GetAllMonitors() []Monitor {
	var monitors []Monitor

	err := monitorCollection.Find(bson.M{}).All(&monitors)
	if err != nil {
		logger.Red("monitor", "Error getting monitors from Mongo: %s", err.Error())
	}

	return monitors
}

func GetMonitor(id string) (Monitor, error) {
	var monitor Monitor

	if !bson.IsObjectIdHex(id) {
		return monitor, ErrorInvalidId
	}

	err := monitorCollection.FindId(bson.ObjectIdHex(id)).One(&monitor)
	if err != nil {
		logger.Red("monitor", "Error getting monitors from Mongo: %s", err.Error())
		return monitor, err
	}

	return monitor, nil
}

func UpdateMonitor(mon *Monitor) error {
	broadcastChange("monchange", *mon)

	return monitorCollection.UpdateId(mon.Id, mon)
}

func AddMonitor(mon *Monitor) error {
	mon.Id = bson.NewObjectId()

	broadcastChange("monadd", *mon)

	return monitorCollection.Insert(mon)
}

func DeleteMonitor(id string) error {
	if !bson.IsObjectIdHex(id) {
		return ErrorInvalidId
	}

	broadcastChange("mondelete", id)

	return monitorCollection.RemoveId(bson.ObjectIdHex(id))
}

func Loop(wg sync.WaitGroup) {
	_, err := GetHost("000000000000000000000000")
	if err != nil {
		p, found := plugins.GetTransports()["localtransport"]
		if !found {
			logger.Red("monitor", "localtransport plugin not found")
		}
		host := Host{
			Id:          bson.ObjectIdHex("000000000000000000000000"),
			Name:        "localhost",
			TransportId: "localtransport",
			Transport:   p().(plugins.Transport),
		}
		hostCollection.Insert(host)
		logger.Yellow("monitor", "Added localhost transport with id %s", host.Id.String())
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

				err = UpdateMonitor(&mon)
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
						logger.Green("monitor", "%s: %s", mon.Id.Hex(), mon.Job.AgentId)
					} else {
						logger.Red("monitor", "%s: %s", mon.Id.Hex(), mon.Job.AgentId)
					}
					mon.LastPoints = p
					mon.LastCheck = t
					mon.NextCheck = t.Add(mon.Interval)

					err = UpdateMonitor(&mon)
					if err != nil {
						logger.Red("monitor", "Error updating: %s", err.Error())
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
