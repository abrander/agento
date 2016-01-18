package plugins

import (
	"github.com/influxdata/influxdb/client/v2"
)

type (
	Agent interface {
		Gather(transport Transport) error
		GetPoints() []*client.Point
	}
)
