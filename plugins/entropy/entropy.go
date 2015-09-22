package entropy

import (
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/influxdb/influxdb/client"

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

	contents, err := ioutil.ReadFile("/proc/sys/kernel/random/entropy_avail")

	if err != nil {
		return err
	}

	availableEntropy, err := strconv.ParseInt(strings.TrimSpace(string(contents)), 10, 64)

	*e = Entropy(availableEntropy)

	return err
}

func (h Entropy) GetPoints() []client.Point {
	//FIXME: Add entropy
	return make([]client.Point, 0)
}
