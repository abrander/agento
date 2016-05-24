package core

import (
	"encoding/json"

	"github.com/BurntSushi/toml"

	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/userdb"
)

type (
	// Host represents a configured host.
	Host struct {
		ID        string            `json:"id"`
		AccountID string            `json:"accountId"`
		Name      string            `json:"name"`
		Transport plugins.Transport `json:"transport"`
	}

	// Hostproxy is used to read TOML configuration for a host.
	hostProxy struct {
		ID          string `toml:"-" json:"_id"`
		AccountID   string `toml:"-" json:"accountId"`
		Name        string `toml:"name" json:"name"`
		TransportID string `toml:"transport" json:"transport"`
	}
)

// GetAccountId will implement userdb.Subject.
func (h *Host) GetAccountId() string {
	return h.AccountID
}

// DecodeTOML will try to decode a simple TOML based configuration.
func (h *Host) DecodeTOML(prim toml.Primitive) error {
	var proxy hostProxy

	err := toml.PrimitiveDecode(prim, &proxy)
	if err != nil {
		return err
	}
	transport, err := plugins.GetTransport(proxy.TransportID)
	if err != nil {
		return err
	}

	err = toml.PrimitiveDecode(prim, transport)
	if err != nil {
		return err
	}

	h.ID = RandomString(20)
	h.AccountID = userdb.God.GetAccountId()
	h.Name = proxy.Name
	h.Transport = transport

	return nil
}

// UnmarshalJSON will try to unmarshal JSON to a Host object.
func (h *Host) UnmarshalJSON(data []byte) error {
	var proxy hostProxy

	err := json.Unmarshal(data, &proxy)
	if err != nil {
		return err
	}

	transport, err := plugins.GetTransport(proxy.TransportID)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, transport)
	if err != nil {
		return err
	}

	h.ID = proxy.ID
	h.AccountID = proxy.AccountID
	h.Name = proxy.Name
	h.Transport = transport

	return nil
}
