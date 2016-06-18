package core

import (
	"encoding/json"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/abrander/agento/plugins"
)

type (
	// Host represents a configured host.
	Host struct {
		ID              string                 `toml:"-" json:"id"`
		AccountID       string                 `toml:"-" json:"accountId"`
		Name            string                 `toml:"name" json:"name"`
		TransportID     string                 `toml:"transport" json:"transport"`
		TransportConfig map[string]interface{} `toml:"config" json:"config"`
	}
)

var (
	transportsLock sync.RWMutex
	transports     map[string]plugins.Transport
)

func init() {
	transportsLock.Lock()
	transports = make(map[string]plugins.Transport)
	transportsLock.Unlock()
}

// GetAccountId will implement userdb.Subject.
func (h *Host) GetAccountId() string {
	return h.AccountID
}

// DecodeTOML will try to decode a simple TOML based configuration.
func (h *Host) DecodeTOML(prim toml.Primitive) error {
	err := toml.PrimitiveDecode(prim, h)
	if err != nil {
		return err
	}

	// Read configuration.
	toml.PrimitiveDecode(prim, &h.TransportConfig)

	// Remove known entries. Someone should find a better method.
	delete(h.TransportConfig, "transport")

	return nil
}

// Transport will return a usable transport for this host.
func (h *Host) Transport() plugins.Transport {
	transportsLock.RLock()
	transport, found := transports[h.ID]
	transportsLock.RUnlock()

	if !found {
		var err error

		transport, err = plugins.GetTransport(h.TransportID)
		if err != nil {
			panic(err.Error())
		}

		// Use JSON as an intermediary for setting configuration. Its ugly,
		// but it does the job for now.
		j, _ := json.Marshal(h.TransportConfig)
		json.Unmarshal(j, transport)

		transportsLock.Lock()
		transports[h.ID] = transport
		transportsLock.Unlock()
	}

	return transport
}
