package core

import (
	"time"

	"github.com/BurntSushi/toml"
	"github.com/influxdata/influxdb/client/v2"

	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/userdb"
)

type (
	// Probe describes a probe measuring something with an agent though a transport.
	Probe struct {
		ID         string          `json:"id"`
		AccountID  string          `json:"accountId"`
		Host       *Host           `json:"host"`
		Interval   time.Duration   `json:"interval"`
		Agent      plugins.Agent   `json:"agent"`
		LastCheck  time.Time       `json:"lastCheck"`
		NextCheck  time.Time       `json:"nextCheck"`
		LastPoints []*client.Point `json:"lastResult"`
	}

	// probeProxy is used to read TOML configuration for a probe.
	probeProxy struct {
		AgentID  string `toml:"agent"`
		Interval int    `toml:"interval"`
		HostID   string `toml:"host"`
	}
)

// GetAccountId will implement userdb.Subject.
func (p *Probe) GetAccountId() string {
	return p.AccountID
}

// DecodeTOML tries to decode a TOML configuration for a probe.
func (p *Probe) DecodeTOML(hostStore HostStore, prim toml.Primitive) error {
	proxy := probeProxy{
		Interval: 10,
	}

	err := toml.PrimitiveDecode(prim, &proxy)
	if err != nil {
		return err
	}

	agent, err := plugins.GetAgent(proxy.AgentID)
	if err != nil {
		return err
	}

	err = toml.PrimitiveDecode(prim, agent)
	if err != nil {
		return err
	}

	host, err := hostStore.GetHost(userdb.God, proxy.HostID)
	if err != nil {
		return err
	}

	p.ID = RandomString(20)
	p.AccountID = userdb.God.GetAccountId()
	p.Agent = agent
	p.Host = host
	p.Interval = time.Duration(proxy.Interval) * time.Second

	return nil
}
