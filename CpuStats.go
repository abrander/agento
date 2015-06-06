package agento

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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

			key := "cpu." + data[0][3:]

			if data[0] == "cpu" {
				key = data[0]
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

func (c *CpuStats) GetMap(m map[string]interface{}) {
	m["misc.Interrupts"] = c.Interrupts
	m["misc.ContextSwitches"] = c.ContextSwitches
	m["misc.Forks"] = c.Forks
	m["misc.RunningProcesses"] = c.RunningProcesses
	m["misc.BlockedProcesses"] = c.BlockedProcesses

	for key, value := range c.Cpu {
		m[key+".User"] = value.User
		m[key+".Nice"] = value.Nice
		m[key+".System"] = value.System
		m[key+".Idle"] = value.Idle
		m[key+".IoWait"] = value.IoWait
		m[key+".Irq"] = value.Irq
		m[key+".SoftIrq"] = value.SoftIrq
		m[key+".Steal"] = value.Steal
		m[key+".Guest"] = value.Guest
		m[key+".GuestNice"] = value.GuestNice
	}
}

func (c *CpuStats) GetDoc(m map[string]string) {
	m["misc.Interrupts"] = "Number of interrupts per second (/s)"
	m["misc.ContextSwitches"] = "Number of context switches per second (/s)"
	m["misc.Forks"] = "Number of forks per second (/s)"
	m["misc.RunningProcesses"] = "Currently running processes (n)"
	m["misc.BlockedProcesses"] = "Number of processes currently blocked (n)"

	m["cpu.<n>.User"] = "Time spend in user mode (ticks/s)"
	m["cpu.<n>.Nice"] = "Time spend in user mode with low priority (ticks/s)"
	m["cpu.<n>.System"] = "Time spend in kernel mode (ticks/s)"
	m["cpu.<n>.Idle"] = "Time spend idle (ticks/s)"
	m["cpu.<n>.IoWait"] = "Time spend waiting for IO (ticks/s)"
	m["cpu.<n>.Irq"] = "Time spend processing interrupts (ticks/s)"
	m["cpu.<n>.SoftIrq"] = "Time spend processing soft interrupts (ticks/s)"
	m["cpu.<n>.Steal"] = "Time spend waiting for the *physical* CPU on a guest (ticks/s)"
	m["cpu.<n>.Guest"] = "Time spend on running guests (ticks/s)"
	m["cpu.<n>.GuestNice"] = "Time spend on running nice guests (ticks/s)"
}
