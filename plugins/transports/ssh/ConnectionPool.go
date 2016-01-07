package ssh

import (
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/abrander/agento/logger"
)

type (
	ConnectionPool struct {
		lock sync.Mutex
		pool map[Ssh]*connection
	}

	connection struct {
		lastUse  time.Time
		client   *ssh.Client
		refCount int
	}
)

var (
	pool ConnectionPool
)

func init() {
	pool.pool = make(map[Ssh]*connection)

	go loop()
}

func loop() {
	ticker := time.Tick(time.Second)
	for t := range ticker {
		pool.lock.Lock()
		for s, conn := range pool.pool {
			if t.Sub(conn.lastUse) > time.Second*10 && conn.refCount == 0 && conn.client != nil {
				conn.client.Close()
				conn.client = nil
				logger.Yellow("ssh", "Closing unused connection %s:%d", s.Host, s.Port)
			}
		}

		pool.lock.Unlock()
	}
}

func (pool *ConnectionPool) Get(ssh Ssh) (*ssh.Client, error) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	conn, found := pool.pool[ssh]
	if found && conn.client != nil {
		conn.refCount++
		return conn.client, nil
	}

	client, err := ssh.Connect()
	if err != nil {
		return nil, err
	}

	pool.pool[ssh] = &connection{client: client, lastUse: time.Now(), refCount: 1}

	return client, nil
}

func (pool *ConnectionPool) Done(ssh Ssh) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	conn, found := pool.pool[ssh]
	if found && conn.client != nil {
		conn.lastUse = time.Now()
		conn.refCount--
	}
}
