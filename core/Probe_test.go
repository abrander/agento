package core

import (
	"testing"

	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/plugins/transports/local"
	"github.com/abrander/agento/userdb"
)

type (
	mockHostStore struct{}
)

func (s *mockHostStore) GetAllHosts(subject userdb.Subject, accountID string) ([]Host, error) {
	return nil, nil
}

func (s *mockHostStore) AddHost(subject userdb.Subject, host *Host) error {
	return nil
}

func (s *mockHostStore) GetHost(subject userdb.Subject, id string) (*Host, error) {
	return &Host{
		Name:      "localhost",
		Transport: localtransport.NewLocalTransport().(plugins.Transport),
	}, nil
}

func (s *mockHostStore) GetHostByName(subject userdb.Subject, name string) (*Host, error) {
	return nil, nil
}

func (s *mockHostStore) DeleteHost(subject userdb.Subject, id string) error {
	return nil
}

func TestProbeDecodeTOML(t *testing.T) {

}
