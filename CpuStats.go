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
	RunningProcesses float64                   `json:"ru"` // Since 2.5.45
	BlockedProcesses float64                   `json:"bl"` // Since 2.5.45
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
			stat.Cpu[data[0]] = &s
		}

		switch data[0] {
		case "intr":
			stat.Interrupts, _ = strconv.ParseFloat(data[1], 64)
		case "ctxt":
			stat.ContextSwitches, _ = strconv.ParseFloat(data[1], 64)
		case "processes":
			stat.Forks, _ = strconv.ParseFloat(data[1], 64)
		case "procs_running":
			stat.RunningProcesses, _ = strconv.ParseFloat(data[1], 64)
		case "procs_blocked":
			stat.BlockedProcesses, _ = strconv.ParseFloat(data[1], 64)
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

func (c *CpuStats) GetMap(m map[string]float64) {
	m["misc.interrupts"] = c.Interrupts
	m["misc.contextswitches"] = c.ContextSwitches
	m["misc.forks"] = c.Forks
	m["misc.runningprocesses"] = c.RunningProcesses
	m["misc.blockedprocesses"] = c.BlockedProcesses

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
