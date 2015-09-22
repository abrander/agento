package hostname

import (
	"os"

	"github.com/influxdb/influxdb/client"

	"github.com/abrander/agento/plugins"
)

func init() {
	plugins.Register("h", NewHostname)
}

type Hostname string

func NewHostname() plugins.Plugin {
	return new(Hostname)
}

func (h *Hostname) Gather() error {
	hostname, err := os.Hostname()
	*h = Hostname(hostname)

	return err
}

func (h Hostname) GetPoints() []client.Point {
	return make([]client.Point, 0)
}
