package agento

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/influxdb/influxdb/client"
)

var previous *CpuStats

type CpuStats struct {
	sampletime       time.Time                 `json:"-"`
	Cpu              map[string]*SingleCpuStat `json:"cpu"`
	Interrupts       float64                   `json:"in"`
	ContextSwitches  float64                   `json:"ct"`
	Forks            float64                   `json:"pr"`
	RunningProcesses int64                     `json:"ru"` // Since 2.5.45
	BlockedProcesses int64                     `json:"bl"` // Since 2.5.45
}

func GetCpuStats() *CpuStats {
	stat := CpuStats{}
	stat.Cpu = make(map[string]*SingleCpuStat)

	path := filepath.Join("/proc/stat")
	file, err := os.Open(path)
	if err != nil {
		return &stat
	}
	defer file.Close()

	stat.sampletime = time.Now()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()

		data := strings.Fields(strings.Trim(text, " "))
		if len(data) < 2 {
			continue
		}

		// cpu* lines
		if strings.HasPrefix(data[0], "cpu") {
			s := SingleCpuStat{}
			s.ReadArray(data)

			key := data[0][3:]

			if data[0] == "cpu" {
				key = "all"
			}

			stat.Cpu[key] = &s
		}

		switch data[0] {
		case "intr":
			stat.Interrupts, _ = strconv.ParseFloat(data[1], 64)
		case "ctxt":
			stat.ContextSwitches, _ = strconv.ParseFloat(data[1], 64)
		case "processes":
			stat.Forks, _ = strconv.ParseFloat(data[1], 64)
		case "procs_running":
			stat.RunningProcesses, _ = strconv.ParseInt(data[1], 10, 64)
		case "procs_blocked":
			stat.BlockedProcesses, _ = strconv.ParseInt(data[1], 10, 64)
		}
	}

	ret := stat.Sub(previous)
	previous = &stat

	return ret
}

func (c *CpuStats) Sub(previous *CpuStats) *CpuStats {
	if previous == nil {
		return &CpuStats{}
	}

	diff := CpuStats{}
	diff.Cpu = make(map[string]*SingleCpuStat)

	duration := float64(c.sampletime.Sub(previous.sampletime)) / float64(time.Second)
	for key, value := range c.Cpu {
		diff.Cpu[key] = value.Sub(previous.Cpu[key], duration)
	}

	diff.Interrupts = Round((c.Interrupts-previous.Interrupts)/duration, 1)
	diff.ContextSwitches = Round((c.ContextSwitches-previous.ContextSwitches)/duration, 1)
	diff.Forks = Round((c.Forks-previous.Forks)/duration, 1)

	// These are not accumulated
	diff.RunningProcesses = c.RunningProcesses
	diff.BlockedProcesses = c.BlockedProcesses

	return &diff
}

func (c *CpuStats) GetPoints() []client.Point {
	points := make([]client.Point, 5+len(c.Cpu)*10)

	points[0] = SimplePoint("misc.Interrupts", c.Interrupts)
	points[1] = SimplePoint("misc.ContextSwitches", c.ContextSwitches)
	points[2] = SimplePoint("misc.Forks", c.Forks)
	points[3] = SimplePoint("misc.RunningProcesses", c.RunningProcesses)
	points[4] = SimplePoint("misc.BlockedProcesses", c.BlockedProcesses)

	i := 5
	for key, value := range c.Cpu {
		points[i+0] = PointWithTag("cpu.User", value.User, "core", key)
		points[i+1] = PointWithTag("cpu.Nice", value.Nice, "core", key)
		points[i+2] = PointWithTag("cpu.System", value.System, "core", key)
		points[i+3] = PointWithTag("cpu.Idle", value.Idle, "core", key)
		points[i+4] = PointWithTag("cpu.IoWait", value.IoWait, "core", key)
		points[i+5] = PointWithTag("cpu.Irq", value.Irq, "core", key)
		points[i+6] = PointWithTag("cpu.SoftIrq", value.SoftIrq, "core", key)
		points[i+7] = PointWithTag("cpu.Steal", value.Steal, "core", key)
		points[i+8] = PointWithTag("cpu.Guest", value.Guest, "core", key)
		points[i+9] = PointWithTag("cpu.GuestNice", value.GuestNice, "core", key)

		i = i + 10
	}

	return points
}

func (c *CpuStats) GetDoc(m map[string]string) {
	m["misc.Interrupts"] = "Number of interrupts per second (/s)"
	m["misc.ContextSwitches"] = "Number of context switches per second (/s)"
	m["misc.Forks"] = "Number of forks per second (/s)"
	m["misc.RunningProcesses"] = "Currently running processes (n)"
	m["misc.BlockedProcesses"] = "Number of processes currently blocked (n)"

	m["cpu.User"] = "Time spend in user mode (ticks/s)"
	m["cpu.Nice"] = "Time spend in user mode with low priority (ticks/s)"
	m["cpu.System"] = "Time spend in kernel mode (ticks/s)"
	m["cpu.Idle"] = "Time spend idle (ticks/s)"
	m["cpu.IoWait"] = "Time spend waiting for IO (ticks/s)"
	m["cpu.Irq"] = "Time spend processing interrupts (ticks/s)"
	m["cpu.SoftIrq"] = "Time spend processing soft interrupts (ticks/s)"
	m["cpu.Steal"] = "Time spend waiting for the *physical* CPU on a guest (ticks/s)"
	m["cpu.Guest"] = "Time spend on running guests (ticks/s)"
	m["cpu.GuestNice"] = "Time spend on running nice guests (ticks/s)"
}
