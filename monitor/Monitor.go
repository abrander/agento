package monitor

import (
	"errors"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/abrander/agento/userdb"
	"github.com/influxdata/influxdb/client/v2"
)

type (
	// Store is an interface describing a complete Agento storage.
	Store interface {
		MonitorStore
		HostStore
	}

	// MonitorStore describes a store capable of storing monitors.
	MonitorStore interface {
		GetAllMonitors(subject userdb.Subject, accountId string) ([]Monitor, error)
		AddMonitor(subject userdb.Subject, mon *Monitor) error
		GetMonitor(subject userdb.Subject, id string) (*Monitor, error)
		UpdateMonitor(subject userdb.Subject, mon *Monitor) error
		DeleteMonitor(subject userdb.Subject, id string) error
	}

	// HostStore is an interface describing a store for hosts.
	HostStore interface {
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

	// Scheduler is a scheduler executing monitor jobs.
	Scheduler struct {
		store   Store
		subject userdb.Subject
	}
)

var (
	ErrorInvalidId error = errors.New("Invalid id")
)

func (m *Monitor) GetAccountId() string {
	return m.AccountId.Hex()
}
