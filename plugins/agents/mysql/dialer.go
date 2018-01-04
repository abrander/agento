package mysql

import (
	"database/sql"
	"net"
	"sync"

	"github.com/go-sql-driver/mysql"

	"github.com/abrander/agento/plugins"
)

var (
	// We have to use a global lock for multiple reasons:
	// 1. mysql.RegisterDial() is not thread safe.
	// 2. Since we have no way of passing a transport to the dialer function
	//    (at open-time), we need to register a new for each dial.
	// 3. We need to assign a dummy dialer after sql.Open() returns to clear
	//    the reference to the transport and to alert other users of mysql.
	// 4. We need to make sure that no one else uses our dialer function.
	dialLock sync.Mutex
)

func init() {
	// We need to register a panicing dial function. Other users of the mysql
	// driver will experience problems. In this way at least they will have a
	// chance of discovering what's wrong.
	mysql.RegisterDial("tcp", panicDialer)
	mysql.RegisterDial("unix", panicDialer)
}

func panicDialer(addr string) (net.Conn, error) {
	panic("The mysql driver is used outside the mysql agent. Fix the mysql driver, or extend our horrible hack")
}

// Dial is a helper function to dial a MySQL server through a Transport. It is
// exposed from this package to allow other packages to use a MySQL connection
// while preserving our ugly global dialer-state-hack.
func Dial(transport plugins.Transport, dsn string) (*sql.DB, error) {
	dialLock.Lock()

	mysql.RegisterDial("tcp", func(addr string) (net.Conn, error) {
		conn, err := transport.Dial("tcp", addr)

		mysql.RegisterDial("tcp", panicDialer)

		dialLock.Unlock()

		return conn, err
	})

	mysql.RegisterDial("unix", func(addr string) (net.Conn, error) {
		conn, err := transport.Dial("unix", addr)

		mysql.RegisterDial("unix", panicDialer)

		dialLock.Unlock()

		return conn, err
	})

	return sql.Open("mysql", dsn)
}
