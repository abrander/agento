package plugins

import (
	"io"
	"net"
)

type (
	Transport interface {
		Plugin
		Dial(network string, address string) (net.Conn, error)
		Exec(cmd string, arguments ...string) (io.Reader, io.Reader, error)
		Open(path string) (io.ReadCloser, error)
		ReadFile(path string) ([]byte, error)
	}
)
