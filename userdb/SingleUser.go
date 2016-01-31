package userdb

import (
	"errors"
)

type (
	// This implements Subject, User, Account and Database for a single user system.
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

func (s *SingleUser) GetAccountId() string {
	return s.GetId()
}

func (s *SingleUser) ResolveKey(key string) (Subject, error) {
	if key == s.key {
		return s, nil

	}

	return nil, errors.New("Wrong key")
}

// This is only here to satisfy the Database interface. This doesn't work
// in singleuser mode. Will always return an error.
func (s *SingleUser) ResolveCookie(value string) (User, error) {
	return nil, errors.New("Cookie auth not supported")
}

func (s *SingleUser) GetAccounts() ([]Account, error) {
	return []Account{s}, nil
}

// Will give access to everything.
func (s *SingleUser) CanAccess(object Object) error {
	return nil
}

// This doesn't do anything in singleuser mode.
func (s *SingleUser) Save() error {
	return nil
}

func (s *SingleUser) GetUsers() ([]User, error) {
	return []User{s}, nil
}
