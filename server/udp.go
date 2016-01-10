package server

import (
	"fmt"
	"sort"

	"github.com/influxdb/influxdb/client"
	"github.com/rcrowley/go-metrics"
)

func init() {
	inventory = make(map[string]*Inventory)
}

type Sample struct {
	Type        int               `json:"t"`
	Probability float64           `json:"p"`
	Identifier  string            `json:"i"`
	Tags        map[string]string `json:"T"`
	Value       float64           `json:"v"`
}

type Inventory struct {
	Histogram  metrics.Histogram
	Identifier string
	Tags       map[string]string
}

var inventory map[string]*Inventory

func (s *Sample) computeKey() string {
	sortedKeys := make([]string, len(s.Tags))

	i := 0
	for key, _ := range s.Tags {
		sortedKeys[i] = key
		i++
	}

	sort.Strings(sortedKeys)

	tags := ""
	for _, tag := range sortedKeys {
		tags = tags + tag + "=" + s.Tags[tag]
	}

	return fmt.Sprintf("%d:%s:%s", s.Type, s.Identifier, tags)
}

func AddUdpSample(s *Sample) error {
	key := s.computeKey()
	intValue := int64(s.Value * 1000000.0)

	i, found := inventory[key]
	if !found {
		hist := metrics.GetOrRegisterHistogram(key, metrics.DefaultRegistry, metrics.NewUniformSample(1001))

		i = &Inventory{
			Histogram:  hist,
			Identifier: s.Identifier,
			Tags:       s.Tags,
		}

		inventory[key] = i
	}

	switch s.Type {
	case 1:
		i.Histogram.Update(intValue)
	}

	return nil
}

func ReportToInfluxdb() {
	for key, value := range inventory {
		// If the histogram was unused for a cycle, we remove it from inventory
		if value.Histogram.Count() == 0 {
			delete(inventory, key)
			continue
		}

		points := make([]client.Point, 1)
		points[0] = client.Point{
			Measurement: value.Identifier,
			Tags:        value.Tags,
			Fields: map[string]interface{}{
				"min":  float64(value.Histogram.Min()) / 1000000.0,
				"max":  float64(value.Histogram.Max()) / 1000000.0,
				"mean": float64(value.Histogram.Mean()) / 1000000.0,
				"p99":  float64(value.Histogram.Percentile(0.99) / 1000000.0),
				"p90":  float64(value.Histogram.Percentile(0.90) / 1000000.0),
			},
		}
		value.Histogram.Sample().Clear()

		WritePoints(points)
	}
}
