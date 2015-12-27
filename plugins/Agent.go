package plugins

import (
	"github.com/influxdb/influxdb/client"
)

type (
	Agent interface {
		Gather() error
		GetPoints() []client.Point
		GetDoc() *Doc
	}
)
