package server

import (
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strconv"
	"time"

	"github.com/rcrowley/go-metrics"

	"github.com/abrander/agento/logger"
	"github.com/abrander/agento/timeseries"
)

type (
	// Sample is a single sampe received from a client.
	Sample struct {
		Type        int               `json:"t"`
		Probability float64           `json:"p"`
		Identifier  string            `json:"i"`
		Tags        map[string]string `json:"T"`
		Value       float64           `json:"v"`
	}

	inventory struct {
		Histogram  metrics.Histogram
		Identifier string
		Tags       map[string]string
	}
)

const (
	// We receive float values from clients. Internally we store them as
	// integers. We do this by multiplying the input by 'exponent' before
	// committing them to the histogram.
	// Before writing the data, we divide by this exponent.
	exponent = 1000000.0
)

func (s *Sample) computeKey() string {
	sortedKeys := make([]string, len(s.Tags))

	i := 0
	for key := range s.Tags {
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

func (s *Server) addUDPSample(sample *Sample) error {
	key := sample.computeKey()
	intValue := int64(sample.Value * exponent)

	i, found := s.inventory[key]
	if !found {
		hist := metrics.GetOrRegisterHistogram(key, metrics.DefaultRegistry, metrics.NewUniformSample(1001))

		i = &inventory{
			Histogram:  hist,
			Identifier: sample.Identifier,
			Tags:       sample.Tags,
		}

		s.inventory[key] = i
	}

	switch sample.Type {
	case 1:
		i.Histogram.Update(intValue)
	}

	return nil
}

func (s *Server) reportToInfluxdb() {
	for key, value := range s.inventory {
		// If the histogram was unused for a cycle, we remove it from inventory
		if value.Histogram.Count() == 0 {
			delete(s.inventory, key)
			continue
		}

		points := make([]*timeseries.Point, 1)
		points[0] = timeseries.NewPoint(
			value.Identifier,
			value.Tags,
			map[string]interface{}{
				"min":  float64(value.Histogram.Min()) / exponent,
				"max":  float64(value.Histogram.Max()) / exponent,
				"mean": float64(value.Histogram.Mean()) / exponent,
				"p99":  float64(value.Histogram.Percentile(0.99) / exponent),
				"p90":  float64(value.Histogram.Percentile(0.90) / exponent),
			},
		)
		value.Histogram.Sample().Clear()

		s.tsdb.WritePoints(points)
	}
}

// ListenAndServeUDP starts the listener.
func (s *Server) ListenAndServeUDP() {
	samples := make(chan *Sample)

	// UDP reader loop
	go func() {
		addr := s.udp.Bind + ":" + strconv.Itoa(int(s.udp.Port))

		laddr, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			logger.Red("server", "ResolveUDPAddr(%s): %s", addr, err.Error())
			return
		}

		conn, err := net.ListenUDP("udp", laddr)
		if err != nil {
			logger.Red("server", "ListenUDP(%s): %s", addr, err.Error())
			return
		}

		defer conn.Close()

		buf := make([]byte, 65535)

		for {
			var sample Sample
			n, _, err := conn.ReadFromUDP(buf)

			if err == nil && json.Unmarshal(buf[:n], &sample) == nil {
				samples <- &sample
			}
		}
	}()

	c := time.Tick(time.Second * time.Duration(s.udp.Interval))

	// Main loop
	for {
		select {
		case sample := <-samples:
			s.addUDPSample(sample)
		case <-c:
			s.reportToInfluxdb()
		}
	}
}
