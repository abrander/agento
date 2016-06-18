package core

import (
	"encoding/json"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/influxdata/influxdb/client/v2"

	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/userdb"
)

type (
	// Probe describes a probe measuring something with an agent though a transport.
	Probe struct {
		ID          string                 `json:"id"`
		AccountID   string                 `json:"accountId"`
		HostID      string                 `toml:"host" json:"host"`
		Interval    time.Duration          `json:"interval"`
		AgentID     string                 `toml:"agent" json:"agent"`
		AgentConfig map[string]interface{} `json:"config"`
		LastCheck   time.Time              `json:"lastCheck"`
		NextCheck   time.Time              `json:"nextCheck"`
		LastPoints  []*client.Point        `json:"lastResult"`
		Tags        map[string]string      `json:"tags"`
	}
)

// GetAccountId will implement userdb.Subject.
func (p *Probe) GetAccountId() string {
	return p.AccountID
}

// DecodeTOML tries to decode a TOML configuration for a probe.
func (p *Probe) DecodeTOML(hostStore HostStore, prim toml.Primitive) error {
	err := toml.PrimitiveDecode(prim, p)
	if err != nil {
		return err
	}

	err = toml.PrimitiveDecode(prim, &p.AgentConfig)
	if err != nil {
		return err
	}
	delete(p.AgentConfig, "agent")

	p.ID = RandomString(20)
	p.AccountID = userdb.God.GetAccountId()

	if p.Interval == 0 {
		p.Interval = time.Second * 10
	} else {
		p.Interval = time.Second
	}

	return nil
}

// Agent will return the agent for a probe.
func (p *Probe) Agent() plugins.Agent {
	// FIXME: Cache this somehow.
	agent, err := plugins.GetAgent(p.AgentID)

	if err != nil {
		panic(err.Error())
	}

	j, _ := json.Marshal(p.AgentConfig)
	json.Unmarshal(j, agent)

	return agent
}
