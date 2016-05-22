package Tcpport

import (
	"time"

	"github.com/influxdata/influxdb/client/v2"

	"github.com/abrander/agento/plugins"
)

func init() {
	plugins.Register("tcpport", newTcpport)
}

// Tcpport will connect to a tcp port and measure timing.
type Tcpport struct {
	Address         string        `json:"a" description:"The address to connect to (host:port)"`
	ConnectDuration time.Duration `json:"c"`
}

func newTcpport() interface{} {
	return new(Tcpport)
}

// Gather will connect and measure.
func (t *Tcpport) Gather(transport plugins.Transport) error {
	start := time.Now()
	conn, err := transport.Dial("tcp", t.Address)
	if err != nil {
		return err
	}
	t.ConnectDuration = time.Now().Sub(start)

	// It doesn't make sense to measure close timing. Go returns without error
	// before the remote end acks.
	err = conn.Close()
	if err != nil {
		return err
	}

	return nil
}

// GetPoints will return a single point describing timing tagged with the address.
func (t *Tcpport) GetPoints() []*client.Point {
	p := make([]*client.Point, 1)

	p[0] = plugins.PointWithTag("tcpport.ConnectDuration", t.ConnectDuration, "address", t.Address)

	return p
}

// GetDoc explains the returned point from GetPoints().
func (t *Tcpport) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("TCP port probe that measures timing")

	doc.AddTag("address", "The address to connect to (hostname:port)")
	doc.AddMeasurement("tcpport.ConnectDuration", "The time it took to open the connection", "ms")

	return doc
}

// Ensure compliance.
var _ plugins.Agent = (*Tcpport)(nil)
