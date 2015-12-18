package entropy

import (
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdb/influxdb/client"

	"github.com/abrander/agento/configuration"
	"github.com/abrander/agento/plugins"
)

func init() {
	plugins.Register("e", NewEntropy)
}

type Entropy int

func NewEntropy() plugins.Plugin {
	return new(Entropy)
}

func (e *Entropy) Gather() error {
	*e = Entropy(0)

	path := filepath.Join(configuration.ProcPath, "/sys/kernel/random/entropy_avail")
	contents, err := ioutil.ReadFile(path)

	if err != nil {
		return err
	}

	availableEntropy, err := strconv.ParseInt(strings.TrimSpace(string(contents)), 10, 64)

	*e = Entropy(availableEntropy)

	return err
}

func (h Entropy) GetPoints() []client.Point {
	points := make([]client.Point, 1)

	points[0] = plugins.SimplePoint("misc.AvailableEntropy", int(h))

	return points
}

func (h Entropy) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("Entropy")

	doc.AddMeasurement("misc.AvailableEntropy", "Available entropy in the kernel pool", "b")

	return doc
}
