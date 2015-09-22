package cpuspeed

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdb/influxdb/client"

	"github.com/abrander/agento"
	"github.com/abrander/agento/plugins"
)

func init() {
	plugins.Register("f", NewCpuSpeed)
}

func NewCpuSpeed() plugins.Plugin {
	return new(CpuSpeed)
}

type CpuSpeed struct {
	Frequency []int
}

func (c *CpuSpeed) Gather() error {
	files, err := filepath.Glob("/sys/devices/system/cpu/cpu[0-9]*/cpufreq/cpuinfo_max_freq")

	if err != nil {
		return err
	}

	c.Frequency = make([]int, len(files))
	i := 0
	for _, file := range files {
		contents, err := ioutil.ReadFile(file)
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

func (c *CpuSpeed) GetPoints() []client.Point {
	points := make([]client.Point, len(c.Frequency))

	for i, frequency := range c.Frequency {
		points[i] = agento.PointWithTag("cpu.Frequency", frequency, "core", strconv.Itoa(i))
	}

	return points
}

func (c *CpuSpeed) GetDoc(m map[string]string) {

	m["cpu.<n>.Frequency"] = "The current CPU frequency (kHz)"
}
