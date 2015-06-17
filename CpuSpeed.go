package agento

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdb/influxdb/client"
)

type CpuSpeed struct {
	Frequency []int
}

func GetCpuSpeed() *CpuSpeed {
	speed := CpuSpeed{}

	files, err := filepath.Glob("/sys/devices/system/cpu/cpu[0-9]*/cpufreq/cpuinfo_max_freq")

	if err != nil {
		return &speed
	}

	speed.Frequency = make([]int, len(files))
	i := 0
	for _, file := range files {
		contents, err := ioutil.ReadFile(file)
		if err == nil {
			speed.Frequency[i], _ = strconv.Atoi(strings.TrimSpace(string(contents)))
			i += 1
		}
	}

	return &speed
}
func (c CpuSpeed) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Frequency)
}

func (c *CpuSpeed) UnmarshalJSON(b []byte) error {
	c.Frequency = make([]int, 100)
	return json.Unmarshal(b, &c.Frequency)
}

func (c *CpuSpeed) GetPoints() []client.Point {
	points := make([]client.Point, len(c.Frequency))

	for i, frequency := range c.Frequency {
		points[i] = PointWithTag("cpu.Frequency", frequency, "core", strconv.Itoa(i))
	}

	return points
}

func (c *CpuSpeed) GetDoc(m map[string]string) {

	m["cpu.<n>.Frequency"] = "The current CPU frequency (kHz)"
}
