package plugins

import (
	"github.com/influxdata/influxdb/client/v2"
)

type (
	// Agent describes the interface an agent must implement. An agent is a
	// collector collecting data and/or metrics from an underlying source.
	Agent interface {
		Gather(transport Transport) error
		GetPoints() []*client.Point
	}
)
