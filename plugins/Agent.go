package plugins

import (
	"github.com/influxdb/influxdb/client"
)

type (
	Agent interface {
		Gather(transport Transport) error
		GetPoints() []client.Point
		GetDoc() *Doc
	}
)
