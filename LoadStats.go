package agento

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdb/influxdb/client"
)

type LoadStats struct {
	Load1       float64 `json:"l1"`
	Load5       float64 `json:"l5"`
	Load15      float64 `json:"l15"`
	ActiveTasks int64   `json:"at"`
	Tasks       int64   `json:"t"`
}

func GetLoadStats() *LoadStats {
	stat := LoadStats{}

	path := filepath.Join("/proc/loadavg")
	file, err := os.Open(path)
	if err != nil {
		return &stat
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()

		data := strings.Fields(strings.Trim(text, " "))
		if len(data) != 5 {
			continue
		}

		stat.Load1, _ = strconv.ParseFloat(data[0], 64)
		stat.Load5, _ = strconv.ParseFloat(data[1], 64)
		stat.Load15, _ = strconv.ParseFloat(data[2], 64)

		sep := strings.Index(data[3], "/")
		if sep > 0 {
			stat.ActiveTasks, _ = strconv.ParseInt(data[3][0:sep], 10, 64)
			stat.Tasks, _ = strconv.ParseInt(data[3][sep+1:], 10, 64)

			// We don't want yo count ourself as active. We're sneeky.
			stat.ActiveTasks -= 1
		}
	}

	return &stat
}

func (l *LoadStats) GetPoints() []client.Point {
	points := make([]client.Point, 5)

	points[0] = SimplePoint("misc.Load1", l.Load1)
	points[1] = SimplePoint("misc.Load5", l.Load5)
	points[2] = SimplePoint("misc.Load15", l.Load15)
	points[3] = SimplePoint("misc.ActiveTasks", l.ActiveTasks)
	points[4] = SimplePoint("misc.Tasks", l.Tasks)

	return points
}

func (l *LoadStats) GetDoc(m map[string]string) {
	m["misc.Load1"] = "System load in the last minute (n)"
	m["misc.Load5"] = "System load in the last 5 minutes (n)"
	m["misc.Load15"] = "System load in the last 15 minutes (n)"
	m["misc.ActiveTasks"] = "Tasks running (n)"
	m["misc.Tasks"] = "Number of tasks"
}
