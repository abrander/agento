package hostname

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/abrander/agento/configuration"
	"github.com/abrander/agento/logger"
	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/timeseries"
)

func init() {
	plugins.Register("hostname", NewHostname)
}

type Hostname string

func NewHostname() interface{} {
	return new(Hostname)
}

func (h *Hostname) Gather(transport plugins.Transport) error {
	hostname := os.Getenv("AGENTO_HOSTNAME")
	hostnamePath := os.Getenv("AGENTO_HOSTNAME_PATH")

	// If we got no hostname from AGENTO_HOSTNAME, try AGENTO_HOSTNAME_PATH
	if hostname == "" && hostnamePath != "" {
		b, err := transport.ReadFile(hostnamePath)
		if err != nil {
			return err
		}

		hostname = strings.TrimSpace(string(b))
	}

	// If we still don't know our hostname, ask proc.
	if hostname == "" {
		path := filepath.Join(configuration.ProcPath, "/sys/kernel/hostname")
		b, err := transport.ReadFile(path)
		if err != nil {
			return err
		}

		hostname = strings.TrimSpace(string(b))
	}

	// If we don't know the hostname by now, something is wrong.
	if hostname == "" {
		logger.Error("config", "Unable to read hostname\n")
	}

	*h = Hostname(hostname)

	return nil
}

func (h Hostname) GetPoints() []*timeseries.Point {
	return make([]*timeseries.Point, 0)
}

func (h Hostname) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("Hostname")

	doc.AddTag("hostname", "The hostname as returned by the hostname command")

	return doc
}

// Ensure compliance
var _ plugins.Agent = (*Hostname)(nil)
