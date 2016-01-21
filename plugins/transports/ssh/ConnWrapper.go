package ssh

// Wrapper to do some form of refcounting on ssh connections after using Dial

import (
	"net"
	"sync"
	"time"

	"github.com/abrander/agento/logger"
)

type (
	ConnWrapper struct {
		sync.Mutex
		underlying net.Conn
		done       bool
		ssh        Ssh
	}
)

func NewConnWrapper(underlying net.Conn, ssh Ssh) net.Conn {
	logger.Green("ssh", "New ConnWrapper allocated for %s", underlying.RemoteAddr().String())

	return &ConnWrapper{
		underlying: underlying,
		ssh:        ssh,
	}
}

func (c *ConnWrapper) Read(b []byte) (n int, err error) {
	return c.underlying.Read(b)
}

func (c *ConnWrapper) Write(b []byte) (n int, err error) {
	return c.underlying.Write(b)
}

func (c *ConnWrapper) Close() error {
	logger.Green("ssh", "ConnWrapper.Close() for %s", c.underlying.RemoteAddr().String())

	c.Lock()
	if !c.done {
		pool.Done(c.ssh)
		c.done = true
	}
	c.Unlock()

	return c.underlying.Close()
}

func (c *ConnWrapper) LocalAddr() net.Addr {
	return c.underlying.LocalAddr()
}

func (c *ConnWrapper) RemoteAddr() net.Addr {
	return c.underlying.RemoteAddr()
}

func (c *ConnWrapper) SetDeadline(t time.Time) error {
	return c.underlying.SetDeadline(t)
}

func (c *ConnWrapper) SetReadDeadline(t time.Time) error {
	return c.underlying.SetReadDeadline(t)
}

func (c *ConnWrapper) SetWriteDeadline(t time.Time) error {
	return c.underlying.SetWriteDeadline(t)
}
