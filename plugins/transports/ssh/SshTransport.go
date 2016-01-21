package ssh

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"syscall"

	"github.com/abrander/agento/logger"
	"github.com/abrander/agento/plugins"
)

type (
	SshTransport struct {
		Ssh
	}
)

func init() {
	plugins.Register("ssh-command", NewSshTransport)
}

func NewSshTransport() plugins.Plugin {
	return new(SshTransport)
}

func (s *SshTransport) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("SSH transport")

	return doc
}

func (s *SshTransport) Exec(cmd string, arguments ...string) (io.Reader, io.Reader, error) {
	for _, arg := range arguments {
		cmd += " " + arg
	}

	logger.Yellow("ssh", "Executing command '%s' on %s:%d as %s", cmd, s.Ssh.Host, s.Ssh.Port, s.Username)
	conn, err := pool.Get(s.Ssh)
	if err != nil {
		return nil, nil, err
	}
	defer pool.Done(s.Ssh)

	session, err := conn.NewSession()
	if err != nil {
		return nil, nil, err
	}
	defer session.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	err = session.Run(cmd)
	if err != nil {
		return &stdoutBuf, &stderrBuf, err
	}

	return &stdoutBuf, &stderrBuf, nil
}

func (s *SshTransport) Dial(network string, address string) (net.Conn, error) {
	conn, err := pool.Get(s.Ssh)
	if err != nil {
		return nil, err
	}

	logger.Yellow("ssh", "Dialing %s://%s via ssh://%s@%s:%d", network, address, s.Ssh.Username, s.Ssh.Host, s.Ssh.Port)

	c, err := conn.Dial(network, address)
	if err != nil {
		return nil, err
	}
	c = NewConnWrapper(c, s.Ssh)

	return c, err
}

func (s *SshTransport) Open(path string) (io.ReadCloser, error) {
	// Maybe "cat" can be replaces by some sftp/scp magic. This should work for now thou.
	r, _, err := s.Exec("/bin/cat", path)

	return ioutil.NopCloser(r), err
}

func (s *SshTransport) ReadFile(path string) ([]byte, error) {
	r, err := s.Open(path)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(r)
}

func (s *SshTransport) Statfs(path string, buf *syscall.Statfs_t) error {
	return errors.New("FIXME: sshtransport does not implement Statfs()")
}

// Ensure compliance
var _ plugins.Transport = (*SshTransport)(nil)
