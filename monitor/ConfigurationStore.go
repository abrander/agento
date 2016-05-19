package monitor

import (
	"errors"
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
	"gopkg.in/mgo.v2/bson"

	"github.com/abrander/agento/configuration"
	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/userdb"
)

type (
	// ConfigurationStore is a read-only store that will read hosts and monitors
	// from the Agento global configuration file.
	ConfigurationStore struct {
		changes  Broadcaster
		metadata toml.MetaData
		hosts    map[bson.ObjectId]Host
		monitors map[bson.ObjectId]Monitor
	}
)

var (
	// ErrNotSupported is returned if an unsupported operation is tried.
	ErrNotSupported = errors.New("Not supported in ConfigurationStore")
)

// NewConfigurationStore will instantiate a new store based on the configuration
// file. This store is read only.
func NewConfigurationStore(config *configuration.Configuration, changes Broadcaster) (*ConfigurationStore, error) {
	s := &ConfigurationStore{
		changes:  changes,
		hosts:    make(map[bson.ObjectId]Host),
		monitors: make(map[bson.ObjectId]Monitor),
	}

	// Retrieve all hosts from configuration.
	meta, primitiveHosts := config.GetAllHosts()

	for name, primitiveHost := range primitiveHosts {
		host := Host{}

		err := meta.PrimitiveDecode(primitiveHost, &host)
		if err != nil {
			return nil, err
		}

		host.Name = name
		if host.Name == "localhost" {
			// localhost is special for now. Scheduler and javascript client
			// expect this id.
			host.Id = bson.ObjectIdHex("000000000000000000000000")
		} else {
			host.Id = bson.NewObjectId()
		}

		// Try to get a transport.
		construct, found := plugins.GetTransports()[host.TransportId]
		if !found {
			return nil, fmt.Errorf("unknown transport '%s'", host.TransportId)
		}

		// Setup the transport.
		transport := construct()
		err = meta.PrimitiveDecode(primitiveHost, transport)
		if err != nil {
			panic(err.Error())
		}

		host.Transport = transport.(plugins.Transport)

		// Save for later.
		s.hosts[host.Id] = host
	}

	// Retrieve monitors from configuration.
	meta, primitiveMonitors := config.GetAllMonitors()

	for _, primitiveMonitor := range primitiveMonitors {
		var proxy struct {
			Host    string `toml:"host"`
			AgentID string `toml:"agent"`
		}

		monitor := Monitor{
			Interval: time.Second * 10,
		}

		err := meta.PrimitiveDecode(primitiveMonitor, &proxy)
		if err != nil {
			return nil, err
		}
		host, err := s.GetHostByName(nil, proxy.Host)
		if err != nil {
			return nil, err
		}

		monitor.HostId = host.Id
		monitor.Id = bson.NewObjectId()

		// Try to find an agent.
		construct, found := plugins.GetAgents()[proxy.AgentID]
		if !found {
			return nil, fmt.Errorf("unknown agent '%s'", proxy.AgentID)
		}

		agent := construct()
		err = meta.PrimitiveDecode(primitiveMonitor, agent)
		if err != nil {
			panic(err.Error())
		}

		monitor.Job.Agent = agent.(plugins.Agent)

		s.monitors[monitor.Id] = monitor
	}

	return s, nil
}

// GetAllHosts returns the complete list of hosts from configuration file.
func (s *ConfigurationStore) GetAllHosts(_ userdb.Subject, _ string) ([]Host, error) {
	l := len(s.hosts)
	hosts := make([]Host, l, l)
	i := 0
	for _, host := range s.hosts {
		hosts[i] = host

		i++
	}

	return hosts, nil
}

// AddHost does nothing in a ConfigurationStore. Will return ErrNotSupported.
func (s *ConfigurationStore) AddHost(_ userdb.Subject, _ *Host) error {
	return ErrNotSupported
}

// GetHost will return the host with the given id.
func (s *ConfigurationStore) GetHost(_ userdb.Subject, id string) (*Host, error) {
	fmt.Printf("LOOKING FOR '%s'\n", id)
	host, found := s.hosts[bson.ObjectIdHex(id)]
	if !found {
		return nil, ErrorInvalidId
	}

	return &host, nil
}

// GetHostByName searches for a host named name.
func (s *ConfigurationStore) GetHostByName(_ userdb.Subject, name string) (*Host, error) {
	for _, host := range s.hosts {
		if host.Name == name {
			return &host, nil
		}
	}

	return nil, fmt.Errorf("Host '%s' not found", name)
}

// DeleteHost does nothing. ConfigurationStore is read-only.
func (s *ConfigurationStore) DeleteHost(_ userdb.Subject, _ string) error {
	return ErrNotSupported
}

// GetAllMonitors return all known monitors.
func (s *ConfigurationStore) GetAllMonitors(_ userdb.Subject, _ string) ([]Monitor, error) {
	l := len(s.monitors)
	monitors := make([]Monitor, l, l)
	i := 0
	for _, monitor := range s.monitors {
		monitors[i] = monitor

		i++
	}

	return monitors, nil
}

// AddMonitor does nothing. ConfigurationStore is read-only.
func (s *ConfigurationStore) AddMonitor(_ userdb.Subject, mon *Monitor) error {
	return ErrNotSupported
}

// GetMonitor will return a monitor identified by id if found.
func (s *ConfigurationStore) GetMonitor(_ userdb.Subject, id string) (*Monitor, error) {
	for _, monitor := range s.monitors {
		if monitor.Id == bson.ObjectIdHex(id) {
			return &monitor, nil
		}
	}

	return nil, fmt.Errorf("Monitor '%s' not found", id)
}

// UpdateMonitor accepts the write but otherwise does no writing to disk.
func (s *ConfigurationStore) UpdateMonitor(_ userdb.Subject, mon *Monitor) error {
	s.monitors[mon.Id] = *mon

	s.changes.Broadcast("monchange", mon)

	return nil
}

// DeleteMonitor does nothing. ConfigurationStore is read-only.
func (s *ConfigurationStore) DeleteMonitor(_ userdb.Subject, _ string) error {
	return ErrNotSupported
}
