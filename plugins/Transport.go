package plugins

import (
	"errors"
	"io"
	"net"
	"syscall"
)

type (
	// Transport defines the interface all transports must implement.
	Transport interface {
		Plugin

		Dial(network string, address string) (net.Conn, error)
		Exec(cmd string, arguments ...string) (io.Reader, io.Reader, error)
		Open(path string) (io.ReadCloser, error)
		ReadFile(path string) ([]byte, error)
		Statfs(path string, buf *syscall.Statfs_t) error
	}
)

// GetTransport will return a transport of type id or nil plus an error if the
// transport was not found.
func GetTransport(id string) (Transport, error) {
	// Try to find a constructor.
	c, found := pluginConstructors[id]
	if !found {
		return nil, errors.New("Transport " + id + " not found")
	}

	// Instantiate.
	plugin := c()

	// Check if the plugin is in fact a transport.
	transport, ok := plugin.(Transport)
	if !ok {
		return nil, errors.New("Transport " + id + " not found")
	}

	return transport, nil
}
