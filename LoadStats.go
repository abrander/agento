package agento

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

func (l *LoadStats) GetMap(m map[string]interface{}) {
	m["misc.load1"] = l.Load1
	m["misc.load5"] = l.Load5
	m["misc.load15"] = l.Load15
	m["misc.activetasks"] = l.ActiveTasks
	m["misc.tasks"] = l.Tasks
}
