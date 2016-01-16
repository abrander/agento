// This implements Subject, User, Account and Database for a single user system
package userdb

import (
	"errors"

	"gopkg.in/mgo.v2/bson"
)

type (
	SingleUser struct {
		key string
	}
)

func NewSingleUser(key string) *SingleUser {
	return &SingleUser{key: key}
}

func (s *SingleUser) GetId() string {
	return "000000000000000000000000"
}

func (s *SingleUser) GetKey() (string, error) {
	return s.key, nil
}

func (s *SingleUser) RefreshKey(key string) error {
	return nil
}

func (s *SingleUser) DeleteKey(key string) error {
	return nil
}

func (s *SingleUser) ResolveKey(key string) (Subject, error) {
	if key == s.key {
		return s, nil

	}

	return nil, errors.New("Wrong key")
}

func (s *SingleUser) ResolveCookie(value string) (User, error) {
	return nil, errors.New("Cookie auth not supported")
}

func (s *SingleUser) GetAccounts() ([]Account, error) {
	return []Account{s}, nil
}

func (s *SingleUser) CanAccess(accountId string) error {
	if !bson.IsObjectIdHex(accountId) {
		return ErrorInvalidAccountId
	}

	return nil
}

func (s *SingleUser) Save() error {
	return nil
}

func (s *SingleUser) GetUsers() ([]User, error) {
	return []User{s}, nil
}
