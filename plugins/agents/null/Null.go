package null

// This is a simple agent demonstrating how an agent should be written and
// structured. It's not really useful for anything else.

import (
	"github.com/influxdata/influxdb/client/v2"

	"github.com/abrander/agento/plugins"
)

func init() {
	plugins.Register("null", newNull)
}

// Null will store the number of bytes read from /dev/null.
type Null struct {
	Length int `json:"n"`
}

func newNull() interface{} {
	return new(Null)
}

// Gather will measure how many bytes can be read from /dev/null.
func (n *Null) Gather(transport plugins.Transport) error {
	b, err := transport.ReadFile("/dev/null")
	if err != nil {
		return err
	}

	n.Length = len(b)

	return nil
}

// GetPoints will return exactly one point. The number of bytes read.
func (n *Null) GetPoints() []*client.Point {
	points := make([]*client.Point, 1)

	points[0] = plugins.SimplePoint("Null.Length", n.Length)

	return points
}

// GetDoc explains the returned points from GetPoints().
func (n *Null) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("null reader (demonstration agent)")

	doc.AddMeasurement("Null.Length", "Number of bytes read from /dev/null", "b")

	return doc
}

// Ensure compliance.
var _ plugins.Agent = (*Null)(nil)
