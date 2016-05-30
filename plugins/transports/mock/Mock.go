package mocktransport

import (
	"bytes"
	"errors"
	"io"
	"net"
	"syscall"

	"github.com/abrander/agento/plugins"
)

func init() {
	plugins.Register("mocktransport", NewMock)
}

func NewMock() interface{} {
	return &Mock{
		files: make(map[string][]byte),
	}
}

type (
	// Mock is a type that can help in writing tests for agents.
	Mock struct {
		files map[string][]byte
	}
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

// SetFile sets a file to be served to an agent if the agent tries to Open() or
// ReadFile() the path.
func (m *Mock) SetFile(path string, contents []byte) {
	m.files[path] = contents
}

func (m *Mock) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("Mock transport for testing")

	return doc
}

func (m *Mock) Dial(network string, address string) (net.Conn, error) {
	return nil, errors.New("Not supported yet")
}

func (m *Mock) Exec(cmd string, arguments ...string) (io.Reader, io.Reader, error) {
	return nil, nil, errors.New("Not supported yet")
}

func (m *Mock) Open(path string) (io.ReadCloser, error) {
	contents, found := m.files[path]
	if !found {
		return nil, errors.New("file not found")
	}

	return nopCloser{bytes.NewBuffer(contents)}, nil
}

func (m *Mock) ReadFile(path string) ([]byte, error) {
	contents, found := m.files[path]
	if !found {
		return nil, errors.New("file not found")
	}

	return contents, nil
}

func (m *Mock) Statfs(path string, buf *syscall.Statfs_t) error {
	return errors.New("Not supported yet")
}

// Ensure compliance
var _ plugins.Transport = (*Mock)(nil)
