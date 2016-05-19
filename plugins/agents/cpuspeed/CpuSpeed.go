package cpuspeed

import (
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdata/influxdb/client/v2"

	"github.com/abrander/agento/configuration"
	"github.com/abrander/agento/plugins"
)

func init() {
	plugins.Register("cpuspeed", NewCpuSpeed)
}

func NewCpuSpeed() interface{} {
	return new(CpuSpeed)
}

type CpuSpeed struct {
	Frequency []int
}

func (c *CpuSpeed) Gather(transport plugins.Transport) error {
	path := filepath.Join(configuration.SysfsPath, "/devices/system/cpu/cpu[0-9]*/cpufreq/cpuinfo_max_freq")
	files, err := filepath.Glob(path)

	if err != nil {
		return err
	}

	c.Frequency = make([]int, len(files))
	i := 0
	for _, file := range files {
		contents, err := transport.ReadFile(file)
		if err == nil {
			c.Frequency[i], _ = strconv.Atoi(strings.TrimSpace(string(contents)))
			i += 1
		}
	}

	return nil
}

func (c CpuSpeed) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Frequency)
}

func (c *CpuSpeed) UnmarshalJSON(b []byte) error {
	c.Frequency = make([]int, 128)
	return json.Unmarshal(b, &c.Frequency)
}

func (c *CpuSpeed) GetPoints() []*client.Point {
	points := make([]*client.Point, len(c.Frequency))

	for i, frequency := range c.Frequency {
		points[i] = plugins.PointWithTag("cpu.Frequency", frequency, "core", strconv.Itoa(i))
	}

	return points
}

func (c *CpuSpeed) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("CPU speed")

	doc.AddTag("core", "The cpu core")

	doc.AddMeasurement("cpu.Frequency", "The current CPU frequency", "kHz")

	return doc
}

// Ensure compliance
var _ plugins.Agent = (*CpuSpeed)(nil)
