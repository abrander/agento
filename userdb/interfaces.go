package userdb

import (
	"errors"
)

type (
	// Object can be manipulated by a Subject (if allowed). Internal
	// objects controlled by users of API's should implement Object.
	Object interface {
		// Get the account id owning the object
		GetAccountId() string
	}

	// A Subject is the 'subject' of an operation.
	Subject interface {
		// Get a alphanumeric id representing a Subject.
		GetId() string

		// Check if the Subject can access an Object. Should return true
		// if allowed, ErrorNoAccess otherwise.
		CanAccess(object Object) error

		// Save the Subject to database.
		Save() error
	}

	// A User is a person using the system.
	User interface {
		Subject

		// Get the accounts the user is connected to.
		GetAccounts() ([]Account, error)
	}

	// An account can represent multiple users.
	Account interface {
		Object
		Subject

		// Get all users associated with the Account.
		GetUsers() ([]User, error)
	}

	Database interface {
		// Resolve a API key to a Subject (can be an User or an Account).
		ResolveKey(key string) (Subject, error)

		// Should map a cookie secret to a User.
		ResolveCookie(value string) (User, error)
	}
)

var (
	// Will be returned from CanAccess() if the Object doesn't have access.
	ErrorNoAccess error = errors.New("access forbidden")

	// Error indicating an invalid account id.
	ErrorInvalidAccountId error = errors.New("invalid account id")

	// Error indicating an invalid user id.
	ErrorInvalidUserId error = errors.New("invalid user id")
)
