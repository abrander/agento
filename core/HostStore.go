package core

import (
	"errors"
	"os"

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

// AddLocalhost will add the magic localhost to the store if needed.
func AddLocalhost(subject userdb.Subject, store HostStore) error {
	hostname, _ := os.Hostname()
	_, err := store.GetHost(subject, "000000000000000000000000")
	if err != nil {
		// Construct the magic host.
		host := &Host{
			ID:          "000000000000000000000000",
			AccountID:   "000000000000000000000000",
			Name:        hostname,
			TransportID: "localtransport",
		}

		// Save it.
		err = store.AddHost(nil, host)
		if err != nil {
			return err
		}
	}

	return nil
}
