package localtransport

import (
	"bytes"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/abrander/agento/plugins"
)

func init() {
	plugins.Register("localtransport", NewLocalTransport)
}

func NewLocalTransport() plugins.Plugin {
	return new(LocalTransport)
}

type (
	LocalTransport struct {
	}
)

func (l *LocalTransport) Dial(network string, address string) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	return dialer.Dial(network, address)
}

func (l *LocalTransport) Exec(cmd string, arguments ...string) (io.Reader, io.Reader, error) {
	command := exec.Command(cmd, arguments...)

	var out bytes.Buffer
	command.Stdout = &out

	stderr, err := command.StderrPipe()
	if err != nil {
		return nil, nil, err
	}

	err = command.Run()

	return &out, stderr, err
}

func (l *LocalTransport) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (l *LocalTransport) ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func (l *LocalTransport) Statfs(path string, buf *syscall.Statfs_t) error {
	return syscall.Statfs(path, buf)
}

// Ensure compliance
var _ plugins.Transport = (*LocalTransport)(nil)
