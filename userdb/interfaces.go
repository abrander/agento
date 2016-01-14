package userdb

import (
	"errors"
)

type (
	Object interface {
		GetAccountId() string
	}

	Subject interface {
		GetId() string
		GetKey() (string, error)
		RefreshKey(key string) error
		DeleteKey(key string) error
		CanAccess(accountId string) error
	}

	User interface {
		Subject
		GetAccounts() ([]Account, error)
	}

	Account interface {
		Subject
		GetUsers() ([]User, error)
	}

	Database interface {
		ResolveKey(key string) (Subject, error)
		ResolveCookie(value string) (User, error)
	}
)

var (
	ErrorNoAccess         error = errors.New("access forbidden")
	ErrorInvalidAccountId error = errors.New("invalid account id")
)
