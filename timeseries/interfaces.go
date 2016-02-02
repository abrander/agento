package timeseries

import (
	"github.com/influxdata/influxdb/client/v2"
)

type (
	Database interface {
		WritePoints(points []*client.Point) error
	}
)
