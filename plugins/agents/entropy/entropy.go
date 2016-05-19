package entropy

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdata/influxdb/client/v2"

	"github.com/abrander/agento/configuration"
	"github.com/abrander/agento/plugins"
)

func init() {
	plugins.Register("entropy", NewEntropy)
}

type Entropy int

func NewEntropy() interface{} {
	return new(Entropy)
}

func (e *Entropy) Gather(transport plugins.Transport) error {
	*e = Entropy(0)

	path := filepath.Join(configuration.ProcPath, "/sys/kernel/random/entropy_avail")
	contents, err := transport.ReadFile(path)

	if err != nil {
		return err
	}

	availableEntropy, err := strconv.ParseInt(strings.TrimSpace(string(contents)), 10, 64)

	*e = Entropy(availableEntropy)

	return err
}

func (h Entropy) GetPoints() []*client.Point {
	points := make([]*client.Point, 1)

	points[0] = plugins.SimplePoint("misc.AvailableEntropy", int(h))

	return points
}

func (h Entropy) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("Entropy")

	doc.AddMeasurement("misc.AvailableEntropy", "Available entropy in the kernel pool", "b")

	return doc
}

// Ensure compliance
var _ plugins.Agent = (*Entropy)(nil)
