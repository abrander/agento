package monitor

import (
	"encoding/json"
	"fmt"

	"gopkg.in/mgo.v2/bson"

	"github.com/abrander/agento/logger"
	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/userdb"
)

type (
	Host struct {
		Id          bson.ObjectId     `json:"id,omitempty" bson:"_id"`
		AccountId   bson.ObjectId     `json:"accountId" bson:"accountId"`
		Name        string            `json:"name"`
		TransportId string            `json:"transportId" bson:"transportId"`
		Transport   plugins.Transport `json:"transport"`
	}
)

func NewHost(name string, transportId string, transport plugins.Transport) *Host {
	return &Host{
		Name:        name,
		TransportId: transportId,
		Transport:   transport,
	}
}

func (s *Scheduler) GetAllHosts(subject userdb.Subject, accountId string) ([]Host, error) {
	var hosts []Host

	err := subject.CanAccess(userdb.ObjectProxy(accountId))
	if err != nil {
		return hosts, err
	}

	err = hostCollection.Find(bson.M{"accountId": bson.ObjectIdHex(accountId)}).All(&hosts)
	if err != nil {
		logger.Red("monitor", "Error getting hosts from Mongo: %s", err.Error())
	}

	return hosts, nil
}

func (s *Scheduler) GetHostByName(subject userdb.Subject, name string) (*Host, error) {
	var host Host

	err := hostCollection.Find(bson.M{"name": name}).One(&host)
	if err != nil {
		return nil, err
	}

	err = subject.CanAccess(&host)
	if err != nil {
		return nil, err
	}

	return &host, nil
}

func (s *Scheduler) GetHost(subject userdb.Subject, id string) (*Host, error) {
	var host Host

	if !bson.IsObjectIdHex(id) {
		return nil, ErrorInvalidId
	}

	err := hostCollection.FindId(bson.ObjectIdHex(id)).One(&host)
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

func (s *Scheduler) AddHost(subject userdb.Subject, host *Host) error {
	host.Id = bson.NewObjectId()

	err := subject.CanAccess(host)
	if err != nil {
		return err
	}

	s.changes.Broadcast("hostadd", host)

	return hostCollection.Insert(host)
}

func (s *Scheduler) DeleteHost(subject userdb.Subject, id string) error {
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

	return hostCollection.RemoveId(bson.ObjectIdHex(id))
}

func (h *Host) GetAccountId() string {
	return h.AccountId.Hex()
}

func (host *Host) UnmarshalJSON(data []byte) error {
	m := make(map[string]json.RawMessage)

	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	idRaw, found := m["id"]
	if found {
		id := ""
		err = json.Unmarshal(idRaw, &id)
		if err != nil {
			return err
		}

		if !bson.IsObjectIdHex(id) {
			return fmt.Errorf("id '%s' is not a valid ObjectId", id)
		}

		host.Id = bson.ObjectIdHex(id)
	}

	accountIdRaw, found := m["accountId"]
	if found {
		id := ""
		err = json.Unmarshal(accountIdRaw, &id)
		if err != nil {
			return err
		}

		if !bson.IsObjectIdHex(id) {
			return fmt.Errorf("accountId '%s' is not a valid ObjectId", id)
		}

		host.AccountId = bson.ObjectIdHex(id)
	}

	nameRaw, found := m["name"]
	if found {
		err = json.Unmarshal(nameRaw, &host.Name)
		if err != nil {
			return err
		}
	}

	agentRaw, found := m["transportId"]
	if !found {
		return fmt.Errorf("transportId not found in document")
	}

	transportRaw, found := m["transport"]
	if !found {
		return fmt.Errorf("transport not found in document")
	}

	err = json.Unmarshal(agentRaw, &host.TransportId)
	if err != nil {
		return err
	}

	a, found := plugins.GetTransports()[host.TransportId]
	if !found {
		return fmt.Errorf("unknown transportId '%s'", host.TransportId)
	}

	transport, ok := a().(plugins.Transport)
	if !ok {
		return fmt.Errorf("plugin '%s' does not implement plugins.transport", host.TransportId)
	}

	host.Transport = transport

	err = json.Unmarshal(transportRaw, host.Transport)
	if err != nil {
		return err
	}

	return nil
}

func (host *Host) SetBSON(raw bson.Raw) error {
	m := make(map[string]bson.Raw)

	err := bson.Unmarshal(raw.Data, &m)
	if err != nil {
		panic(err.Error())
	}

	idRaw, found := m["_id"]
	if found {
		err = idRaw.Unmarshal(&host.Id)
		if err != nil {
			return err
		}
	}

	accountIdRaw, found := m["accountId"]
	if found {
		err = accountIdRaw.Unmarshal(&host.AccountId)
		if err != nil {
			return err
		}
	}

	nameRaw, found := m["name"]
	if found {
		err = nameRaw.Unmarshal(&host.Name)
		if err != nil {
			return err
		}
	}

	transportRaw, found := m["transportId"]
	if !found {
		return fmt.Errorf("transportId not found in document")
	}

	err = transportRaw.Unmarshal(&host.TransportId)
	if err != nil {
		return err
	}

	a, found := plugins.GetTransports()[host.TransportId]
	if !found {
		return fmt.Errorf("unknown transportId '%s'", host.TransportId)
	}

	transport, ok := a().(plugins.Transport)
	if !ok {
		return fmt.Errorf("plugin '%s' does not implement plugins.transport", host.TransportId)
	}

	host.Transport = transport

	transportRaw, found = m["transport"]
	if !found {
		return fmt.Errorf("transport not found in document")
	}

	err = transportRaw.Unmarshal(host.Transport)
	if err != nil {
		return err
	}

	return nil
}
