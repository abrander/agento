package monitor

import (
	"encoding/json"
	"fmt"

	"gopkg.in/mgo.v2/bson"

	"github.com/abrander/agento/plugins"
)

type (
	Host struct {
		Id          bson.ObjectId     `json:"id,omitempty" bson:"_id"`
		AccountId   bson.ObjectId     `json:"accountId" bson:"accountId"`
		Name        string            `json:"name"`
		TransportId string            `toml:"transport" json:"transportId" bson:"transportId"`
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
