package monitor

import (
	"fmt"

	"github.com/BurntSushi/toml"

	"github.com/abrander/agento/configuration"
	"github.com/abrander/agento/core"
	"github.com/abrander/agento/userdb"
)

type (
	// ConfigurationStore is a read-only store that will read hosts and probes
	// from the Agento global configuration file.
	ConfigurationStore struct {
		changes  core.Broadcaster
		metadata toml.MetaData
		hosts    map[string]core.Host
		probes   map[string]core.Probe
	}
)

// NewConfigurationStore will instantiate a new store based on the configuration
// file. This store is read only.
func NewConfigurationStore(config *configuration.Configuration, changes core.Broadcaster) (*ConfigurationStore, error) {
	s := &ConfigurationStore{
		changes: changes,
		hosts:   make(map[string]core.Host),
		probes:  make(map[string]core.Probe),
	}

	// Retrieve all hosts from configuration.
	primitiveHosts := config.GetHostPrimitives()

	for id, primitiveHost := range primitiveHosts {
		host := core.Host{}

		err := host.DecodeTOML(primitiveHost)
		if err != nil {
			return nil, err
		}
		host.ID = id

		// Save for later.
		s.hosts[host.ID] = host
	}

	// Retrieve probes from configuration.
	primitiveProbes := config.GetProbePrimitives()

	for id, primitiveProbe := range primitiveProbes {
		probe := core.Probe{}

		err := probe.DecodeTOML(s, primitiveProbe)
		if err != nil {
			return nil, err
		}

		probe.ID = id

		s.probes[probe.ID] = probe
	}

	return s, nil
}

// GetAllHosts returns the complete list of hosts from configuration file.
func (s *ConfigurationStore) GetAllHosts(_ userdb.Subject, _ string) ([]core.Host, error) {
	l := len(s.hosts)
	hosts := make([]core.Host, l, l)
	i := 0
	for _, host := range s.hosts {
		hosts[i] = host

		i++
	}

	return hosts, nil
}

// AddHost adds a host to memory, not to the configuration file.
func (s *ConfigurationStore) AddHost(_ userdb.Subject, host *core.Host) error {
	host.ID = core.RandomString(20)
	s.hosts[host.ID] = *host

	s.changes.Broadcast("hostadd", host)

	return nil
}

// GetHost will return the host with the given id.
func (s *ConfigurationStore) GetHost(_ userdb.Subject, id string) (*core.Host, error) {
	host, found := s.hosts[id]
	if !found {
		return nil, core.ErrHostNotFound
	}

	return &host, nil
}

// GetHostByName searches for a host named name.
func (s *ConfigurationStore) GetHostByName(_ userdb.Subject, name string) (*core.Host, error) {
	for _, host := range s.hosts {
		if host.Name == name {
			return &host, nil
		}
	}

	return nil, fmt.Errorf("Host '%s' not found", name)
}

// DeleteHost will remove a host from memory, but not from configuration file.
func (s *ConfigurationStore) DeleteHost(_ userdb.Subject, id string) error {
	host, found := s.hosts[id]
	if !found {
		return core.ErrHostNotFound
	}

	delete(s.hosts, id)

	s.changes.Broadcast("hostdelete", &host)

	return nil
}

// GetAllProbes return all known probes.
func (s *ConfigurationStore) GetAllProbes(_ userdb.Subject, _ string) ([]core.Probe, error) {
	l := len(s.probes)
	probes := make([]core.Probe, l, l)
	i := 0
	for _, probe := range s.probes {
		probes[i] = probe

		i++
	}

	return probes, nil
}

// AddProbe adds a probe to memory.
func (s *ConfigurationStore) AddProbe(_ userdb.Subject, probe *core.Probe) error {
	probe.ID = core.RandomString(20)
	s.probes[probe.ID] = *probe

	s.changes.Broadcast("probeadd", probe)

	return nil
}

// GetProbe will return a probe identified by id if found.
func (s *ConfigurationStore) GetProbe(_ userdb.Subject, id string) (*core.Probe, error) {
	probe, found := s.probes[id]
	if !found {
		return nil, core.ErrProbeNotFound
	}

	return &probe, nil
}

// UpdateProbe accepts the write but otherwise does no writing to disk.
func (s *ConfigurationStore) UpdateProbe(_ userdb.Subject, probe *core.Probe) error {
	s.probes[probe.ID] = *probe

	s.changes.Broadcast("probechange", probe)

	return nil
}

// DeleteProbe does delete the probe from memory but not from file.
func (s *ConfigurationStore) DeleteProbe(_ userdb.Subject, id string) error {
	probe, found := s.probes[id]
	if !found {
		return core.ErrProbeNotFound
	}

	delete(s.probes, id)

	s.changes.Broadcast("probedelete", &probe)

	return nil
}
