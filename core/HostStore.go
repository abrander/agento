package core

import (
	"errors"

	"github.com/abrander/agento/userdb"
)

// HostStore is an interface describing a store for hosts.
type HostStore interface {
	GetAllHosts(subject userdb.Subject, accountID string) ([]Host, error)
	AddHost(subject userdb.Subject, host *Host) error
	GetHost(subject userdb.Subject, id string) (*Host, error)
	GetHostByName(subject userdb.Subject, name string) (*Host, error)
	DeleteHost(subject userdb.Subject, id string) error
}

var (
	// ErrHostNotFound will be returned iof the host cannot be found.
	ErrHostNotFound = errors.New("Host not found")
)
