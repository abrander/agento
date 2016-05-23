package monitor

import (
	"os"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/abrander/agento/configuration"
	"github.com/abrander/agento/core"
	"github.com/abrander/agento/logger"
	"github.com/abrander/agento/userdb"
)

type (
	// MongoStore is an implementation of Store using MongoDB as a backend.
	MongoStore struct {
		changes           core.Broadcaster
		sess              *mgo.Session
		db                *mgo.Database
		hostCollection    *mgo.Collection
		monitorCollection *mgo.Collection
	}
)

// NewMongoStore will instantiate a new MongoStore. MongoStore is as the name
// suggest backed by a MongoDB database.
func NewMongoStore(config configuration.MongoConfiguration, changes core.Broadcaster) (*MongoStore, error) {
	var err error
	m := &MongoStore{}

	m.sess, err = mgo.Dial(config.Url)
	if err != nil {
		logger.Error("monitor", "Can't connect to mongo, go error %v", err)
		os.Exit(1)
	}
	logger.Green("monitor", "Connected to mongo/%s at %s", config.Database, config.Url)

	m.db = m.sess.DB(config.Database)
	m.hostCollection = m.db.C("hosts")
	m.monitorCollection = m.db.C("monitors")

	m.changes = changes

	return m, nil
}

// GetAllMonitors will return all monitors belonging to accountID that is
// accessible by subject.
func (s *MongoStore) GetAllMonitors(subject userdb.Subject, accountID string) ([]Monitor, error) {
	var monitors []Monitor

	err := subject.CanAccess(userdb.ObjectProxy(accountID))
	if err != nil {
		return []Monitor{}, err
	}

	err = s.monitorCollection.Find(bson.M{"accountID": bson.ObjectIdHex(accountID)}).All(&monitors)
	if err != nil {
		return []Monitor{}, err
	}

	return monitors, nil
}

// GetMonitor will return the monitor identified by id if accessible by subject.
func (s *MongoStore) GetMonitor(subject userdb.Subject, id string) (*Monitor, error) {
	var monitor Monitor

	if !bson.IsObjectIdHex(id) {
		return &monitor, ErrorInvalidId
	}

	err := s.monitorCollection.FindId(bson.ObjectIdHex(id)).One(&monitor)
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

// UpdateMonitor will save the monitor is allowed by subject.
func (s *MongoStore) UpdateMonitor(subject userdb.Subject, mon *Monitor) error {
	_, err := s.GetMonitor(subject, mon.Id.Hex())
	if err != nil {
		return err
	}

	s.changes.Broadcast("monchange", mon)

	return s.monitorCollection.UpdateId(mon.Id, mon)
}

// AddMonitor will add a new monitor. Everyone can add monitors, but subject
// cannot add a monitor that the subject cannot access itself.
func (s *MongoStore) AddMonitor(subject userdb.Subject, mon *Monitor) error {
	mon.Id = bson.NewObjectId()

	err := subject.CanAccess(mon)
	if err != nil {
		return err
	}

	s.changes.Broadcast("monadd", mon)

	return s.monitorCollection.Insert(mon)
}

// DeleteMonitor does exactly that. Deletes a monitor.
func (s *MongoStore) DeleteMonitor(subject userdb.Subject, id string) error {
	if !bson.IsObjectIdHex(id) {
		return ErrorInvalidId
	}

	mon, err := s.GetMonitor(subject, id)
	if err != nil {
		return err
	}

	s.changes.Broadcast("mondelete", mon)

	return s.monitorCollection.RemoveId(bson.ObjectIdHex(id))
}

// GetAllHosts will return all hosts accessible by subject.
func (s *MongoStore) GetAllHosts(subject userdb.Subject, accountID string) ([]Host, error) {
	var hosts []Host

	err := subject.CanAccess(userdb.ObjectProxy(accountID))
	if err != nil {
		return hosts, err
	}

	err = s.hostCollection.Find(bson.M{"accountID": bson.ObjectIdHex(accountID)}).All(&hosts)
	if err != nil {
		logger.Red("monitor", "Error getting hosts from Mongo: %s", err.Error())
	}

	return hosts, nil
}

// GetHostByName will return the host matching name.
func (s *MongoStore) GetHostByName(subject userdb.Subject, name string) (*Host, error) {
	var host Host

	err := s.hostCollection.Find(bson.M{"name": name}).One(&host)
	if err != nil {
		return nil, err
	}

	err = subject.CanAccess(&host)
	if err != nil {
		return nil, err
	}

	return &host, nil
}

// GetHost returns a host matching id.
func (s *MongoStore) GetHost(subject userdb.Subject, id string) (*Host, error) {
	var host Host

	if !bson.IsObjectIdHex(id) {
		return nil, ErrorInvalidId
	}

	err := s.hostCollection.FindId(bson.ObjectIdHex(id)).One(&host)
	if err != nil {
		logger.Red("host", "Error getting host from Mongo: %s", err.Error())
		return nil, err
	}

	err = subject.CanAccess(&host)
	if err != nil {
		return nil, err
	}

	return &host, nil
}

// AddHost will add a new host. Please note that this will not ensure that host
// is owned by subject, but merely ensure that subject is allowed to create the
// host.
func (s *MongoStore) AddHost(subject userdb.Subject, host *Host) error {
	host.Id = bson.NewObjectId()

	host.AccountId = bson.ObjectIdHex(subject.GetId())
	err := subject.CanAccess(host)
	if err != nil {
		return err
	}

	s.changes.Broadcast("hostadd", host)

	return s.hostCollection.Insert(host)
}

// DeleteHost will delete a host matching id.
func (s *MongoStore) DeleteHost(subject userdb.Subject, id string) error {
	if !bson.IsObjectIdHex(id) {
		return ErrorInvalidId
	}

	host, err := s.GetHost(subject, id)
	if err != nil {
		return err
	}

	err = subject.CanAccess(host)
	if err != nil {
		return err
	}

	s.changes.Broadcast("hostdelete", host)

	return s.hostCollection.RemoveId(bson.ObjectIdHex(id))
}
