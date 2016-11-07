package timeseries

import (
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

type (
	// Point represents one sample.
	Point struct {
		Time   time.Time              `json:"time"`
		Name   string                 `json:"name"`
		Tags   map[string]string      `json:"tags"`
		Fields map[string]interface{} `json:"fields"`
	}
)

func NewPoint(name string, tags map[string]string, fields map[string]interface{}, t ...time.Time) *Point {
	var T time.Time
	if len(t) > 0 {
		T = t[0]
	}

	if tags == nil {
		tags = make(map[string]string)
	}

	if fields == nil {
		fields = make(map[string]interface{})
	}

	return &Point{
		Time:   T,
		Name:   name,
		Tags:   tags,
		Fields: fields,
	}
}

// InfluxDBPoint will return an InfluxDB compatible point.
func (p *Point) InfluxDBPoint() *client.Point {
	point, _ := client.NewPoint(p.Name, p.Tags, p.Fields, p.Time)

	return point
}
