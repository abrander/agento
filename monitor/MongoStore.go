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
		changes         core.Broadcaster
		sess            *mgo.Session
		db              *mgo.Database
		hostCollection  *mgo.Collection
		probeCollection *mgo.Collection
	}
)

// NewMongoStore will instantiate a new MongoStore. MongoStore is as the name
// suggest backed by a MongoDB database.
func NewMongoStore(config configuration.MongoConfiguration, changes core.Broadcaster) (*MongoStore, error) {
	var err error
	m := &MongoStore{}

	m.sess, err = mgo.Dial(config.Url)
	if err != nil {
		logger.Error("mongostore", "Can't connect to mongo, go error %v", err)
		os.Exit(1)
	}
	logger.Green("mongostore", "Connected to mongo/%s at %s", config.Database, config.Url)

	m.db = m.sess.DB(config.Database)
	m.hostCollection = m.db.C("hosts")
	m.probeCollection = m.db.C("probes")

	m.changes = changes

	return m, nil
}

// GetAllProbes will return all probes belonging to accountID that is
// accessible by subject.
func (s *MongoStore) GetAllProbes(subject userdb.Subject, accountID string) ([]core.Probe, error) {
	var probes []core.Probe

	err := subject.CanAccess(userdb.ObjectProxy(accountID))
	if err != nil {
		return nil, err
	}

	err = s.probeCollection.Find(bson.M{"accountID": bson.ObjectIdHex(accountID)}).All(&probes)
	if err != nil {
		return nil, err
	}

	return probes, nil
}

// GetProbe will return the probe identified by id if accessible by subject.
func (s *MongoStore) GetProbe(subject userdb.Subject, id string) (*core.Probe, error) {
	var probe core.Probe

	if !bson.IsObjectIdHex(id) {
		return nil, core.ErrProbeNotFound
	}

	err := s.probeCollection.FindId(bson.ObjectIdHex(id)).One(&probe)
	if err != nil {
		logger.Red("mongostore", "Error getting probe %s from Mongo: %s", id, err.Error())
		return nil, err
	}

	err = subject.CanAccess(&probe)
	if err != nil {
		return nil, err
	}

	return &probe, nil
}

// UpdateProbe will save the probe is allowed by subject.
func (s *MongoStore) UpdateProbe(subject userdb.Subject, probe *core.Probe) error {
	_, err := s.GetProbe(subject, probe.ID)
	if err != nil {
		return err
	}

	s.changes.Broadcast("probechange", probe)

	return s.probeCollection.UpdateId(bson.ObjectIdHex(probe.ID), probe)
}

// AddProbe will add a new probe. Everyone can add probes, but subject
// cannot add a probe that the subject cannot access itself.
func (s *MongoStore) AddProbe(subject userdb.Subject, probe *core.Probe) error {
	probe.ID = bson.NewObjectId().Hex()

	err := subject.CanAccess(probe)
	if err != nil {
		return err
	}

	s.changes.Broadcast("probeadd", probe)

	return s.probeCollection.Insert(probe)
}

// DeleteProbe does exactly that. Deletes a probe.
func (s *MongoStore) DeleteProbe(subject userdb.Subject, id string) error {
	if !bson.IsObjectIdHex(id) {
		return core.ErrProbeNotFound
	}

	probe, err := s.GetProbe(subject, id)
	if err != nil {
		return err
	}

	s.changes.Broadcast("probedelete", probe)

	return s.probeCollection.RemoveId(bson.ObjectIdHex(id))
}

// GetAllHosts will return all hosts accessible by subject.
func (s *MongoStore) GetAllHosts(subject userdb.Subject, accountID string) ([]core.Host, error) {
	var hosts []core.Host

	err := subject.CanAccess(userdb.ObjectProxy(accountID))
	if err != nil {
		return hosts, err
	}

	err = s.hostCollection.Find(bson.M{"accountID": bson.ObjectIdHex(accountID)}).All(&hosts)
	if err != nil {
		logger.Red("mongostore", "Error getting hosts from Mongo: %s", err.Error())
	}

	return hosts, nil
}

// GetHostByName will return the host matching name.
func (s *MongoStore) GetHostByName(subject userdb.Subject, name string) (*core.Host, error) {
	var host core.Host

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
func (s *MongoStore) GetHost(subject userdb.Subject, id string) (*core.Host, error) {
	var host core.Host

	if !bson.IsObjectIdHex(id) {
		return nil, core.ErrHostNotFound
	}

	err := s.hostCollection.FindId(bson.ObjectIdHex(id)).One(&host)
	if err != nil {
		logger.Red("mongostore", "Error getting host from Mongo: %s", err.Error())
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
func (s *MongoStore) AddHost(subject userdb.Subject, host *core.Host) error {
	host.ID = bson.NewObjectId().Hex()

	host.AccountID = subject.GetId()
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
		return core.ErrHostNotFound
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
