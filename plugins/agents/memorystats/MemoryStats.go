package memorystats

import (
	"bufio"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdata/influxdb/client/v2"

	"github.com/abrander/agento/configuration"
	"github.com/abrander/agento/plugins"
)

func init() {
	plugins.Register("memory", NewMemoryStats)
}

func NewMemoryStats() interface{} {
	return new(MemoryStats)
}

type MemoryStats struct {
	Used     int64 `json:"u"`
	Free     int64 `json:"f"`
	Shared   int64 `json:"s"`
	Buffers  int64 `json:"b"`
	Cached   int64 `json:"c"`
	SwapUsed int64 `json:"su"`
	SwapFree int64 `json:"sf"`
}

func getMemInfo(transport plugins.Transport) *map[string]int64 {
	m := make(map[string]int64)

	path := filepath.Join(configuration.ProcPath, "/meminfo")
	file, err := transport.Open(path)
	if err != nil {
		return &m
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()

		n := strings.Index(text, ":")
		if n == -1 {
			continue
		}

		key := text[:n]
		data := strings.Split(strings.Trim(text[(n+1):], " "), " ")
		if len(data) == 1 {
			value, err := strconv.ParseInt(data[0], 10, 64)
			if err != nil {
				continue
			}
			m[key] = value
		} else if len(data) == 2 {
			if data[1] == "kB" {
				value, err := strconv.ParseInt(data[0], 10, 64)
				if err != nil {
					continue
				}

				m[key] = value
			}
		}
	}

	return &m
}

func (stat *MemoryStats) Gather(transport plugins.Transport) error {
	meminfo := getMemInfo(transport)

	stat.Used = (*meminfo)["MemTotal"] - (*meminfo)["MemFree"] - (*meminfo)["Buffers"] - (*meminfo)["Cached"]
	stat.Free = (*meminfo)["MemFree"]
	stat.Shared = (*meminfo)["Shmem"]
	stat.Buffers = (*meminfo)["Buffers"]
	stat.Cached = (*meminfo)["Cached"]

	stat.SwapUsed = (*meminfo)["SwapTotal"] - (*meminfo)["SwapFree"]
	stat.SwapFree = (*meminfo)["SwapFree"]

	return nil
}

func (s *MemoryStats) GetPoints() []*client.Point {
	points := make([]*client.Point, 7)

	points[0] = plugins.SimplePoint("mem.Used", s.Used)
	points[1] = plugins.SimplePoint("mem.Free", s.Free)
	points[2] = plugins.SimplePoint("mem.Shared", s.Shared)
	points[3] = plugins.SimplePoint("mem.Buffers", s.Buffers)
	points[4] = plugins.SimplePoint("mem.Cached", s.Cached)
	points[5] = plugins.SimplePoint("swap.Used", s.SwapUsed)
	points[6] = plugins.SimplePoint("swap.Free", s.SwapFree)

	return points
}

func (s *MemoryStats) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("Memory & Swap")

	doc.AddMeasurement("mem.Used", "Memory used", "b")
	doc.AddMeasurement("mem.Free", "Free memory", "b")
	doc.AddMeasurement("mem.Shared", "Memory shared among multiple processes", "b")
	doc.AddMeasurement("mem.Buffers", "Memory used for buffers", "b")
	doc.AddMeasurement("mem.Cached", "Memory used for cache", "b")
	doc.AddMeasurement("swap.Used", "Used swap", "b")
	doc.AddMeasurement("swap.Free", "Free swap", "b")

	return doc
}

// Ensure compliance
var _ plugins.Agent = (*MemoryStats)(nil)
