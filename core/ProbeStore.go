package core

import (
	"errors"

	"github.com/abrander/agento/userdb"
)

// ProbeStore describes a store capable of storing probes.
type ProbeStore interface {
	GetAllProbes(subject userdb.Subject, accountID string) ([]Probe, error)
	AddProbe(subject userdb.Subject, probe *Probe) error
	GetProbe(subject userdb.Subject, id string) (*Probe, error)
	UpdateProbe(subject userdb.Subject, probe *Probe) error
	DeleteProbe(subject userdb.Subject, id string) error
}

var (
	// ErrProbeNotFound will be returned if the probe cannot be found.
	ErrProbeNotFound = errors.New("Probe not found")
)
