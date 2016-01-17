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
		CanAccess(object Object) error
		Save() error
	}

	User interface {
		Subject
		GetAccounts() ([]Account, error)
	}

	Account interface {
		Object
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
	ErrorInvalidUserId    error = errors.New("invalid user id")
)
