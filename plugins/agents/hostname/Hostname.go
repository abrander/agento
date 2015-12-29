package hostname

import (
	"path/filepath"
	"strings"

	"github.com/influxdb/influxdb/client"

	"github.com/abrander/agento/configuration"
	"github.com/abrander/agento/plugins"
)

func init() {
	plugins.Register("h", NewHostname)
}

type Hostname string

func NewHostname() plugins.Plugin {
	return new(Hostname)
}

func (h *Hostname) Gather(transport plugins.Transport) error {
	path := filepath.Join(configuration.ProcPath, "/sys/kernel/hostname")
	b, err := transport.ReadFile(path)
	if err != nil {
		return err
	}

	hostname := strings.TrimSpace(string(b))
	*h = Hostname(hostname)

	return err
}

func (h Hostname) GetPoints() []client.Point {
	return make([]client.Point, 0)
}

func (h Hostname) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("Hostname")

	doc.AddTag("hostname", "The hostname as returned by the hostname command")

	return doc
}

// Ensure compliance
var _ plugins.Agent = (*Hostname)(nil)
