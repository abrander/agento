package core

type (
	// Store is an interface describing a complete Agento storage.
	Store interface {
		HostStore
		ProbeStore
	}
)
