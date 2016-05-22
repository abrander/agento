package netfilter

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/influxdata/influxdb/client/v2"

	"github.com/abrander/agento/configuration"
	"github.com/abrander/agento/plugins"
)

func init() {
	plugins.Register("netfilter", newNetfilter)
}

func newNetfilter() interface{} {
	return new(Netfilter)
}

// Netfilter will collect information about the Linux kernel Netfilter.
type Netfilter struct {
	ConnTrackCount int64 `json:"c"`
}

// Gather will read stats from /proc
func (n *Netfilter) Gather(transport plugins.Transport) error {
	path := filepath.Join(configuration.ProcPath, "/sys/net/netfilter/nf_conntrack_count")
	contents, err := transport.ReadFile(path)

	// If the file doesn't exist, we assume that netfilter is not tracking
	// connections. We set the value to -1 and abort.
	if p, ok := err.(*os.PathError); ok && p.Err == syscall.ENOENT {
		n.ConnTrackCount = -1

		return nil
	}

	if err != nil {
		fmt.Printf("%T %+v\n", err, err)
		return err
	}

	n.ConnTrackCount, err = strconv.ParseInt(string(contents), 10, 64)
	if err != nil {
		return err
	}

	return nil
}

// GetPoints will return a single point for now.
func (n *Netfilter) GetPoints() []*client.Point {
	points := make([]*client.Point, 1)

	points[0] = plugins.SimplePoint("netfilter.ConnectionsTracked", n.ConnTrackCount)

	return points
}

// GetDoc tries to explain our single point.
func (n *Netfilter) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("Netfilter usage")

	doc.AddMeasurement("netfilter.ConnectionsTracked",
		"The number currently tracked connections (or -1 if tracking is disabled)",
		"n")

	return doc
}

// Ensure compliance
var _ plugins.Agent = (*Netfilter)(nil)
