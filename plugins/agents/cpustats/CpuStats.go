package cpustats

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
	plugins.Register("cpustats", NewCpuStats)
}

type CpuStats struct {
	Cpu              map[string]*SingleCpuStat `json:"cpu"`
	Interrupts       float64                   `json:"in"`
	ContextSwitches  float64                   `json:"ct"`
	Forks            float64                   `json:"pr"`
	RunningProcesses int64                     `json:"ru"` // Since 2.5.45
	BlockedProcesses int64                     `json:"bl"` // Since 2.5.45
}

func NewCpuStats() plugins.Plugin {
	return new(CpuStats)
}

func (stat *CpuStats) Gather(transport plugins.Transport) error {
	stat.Cpu = make(map[string]*SingleCpuStat)

	path := filepath.Join(configuration.ProcPath, "/stat")
	file, err := transport.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

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

	return nil
}

func (c *CpuStats) GetPoints() []*client.Point {
	points := make([]*client.Point, 5+len(c.Cpu)*10)

	points[0] = plugins.SimplePoint("misc.Interrupts", c.Interrupts)
	points[1] = plugins.SimplePoint("misc.ContextSwitches", c.ContextSwitches)
	points[2] = plugins.SimplePoint("misc.Forks", c.Forks)
	points[3] = plugins.SimplePoint("misc.RunningProcesses", c.RunningProcesses)
	points[4] = plugins.SimplePoint("misc.BlockedProcesses", c.BlockedProcesses)

	i := 5
	for key, value := range c.Cpu {
		points[i+0] = plugins.PointWithTag("cpu.User", value.User, "core", key)
		points[i+1] = plugins.PointWithTag("cpu.Nice", value.Nice, "core", key)
		points[i+2] = plugins.PointWithTag("cpu.System", value.System, "core", key)
		points[i+3] = plugins.PointWithTag("cpu.Idle", value.Idle, "core", key)
		points[i+4] = plugins.PointWithTag("cpu.IoWait", value.IoWait, "core", key)
		points[i+5] = plugins.PointWithTag("cpu.Irq", value.Irq, "core", key)
		points[i+6] = plugins.PointWithTag("cpu.SoftIrq", value.SoftIrq, "core", key)
		points[i+7] = plugins.PointWithTag("cpu.Steal", value.Steal, "core", key)
		points[i+8] = plugins.PointWithTag("cpu.Guest", value.Guest, "core", key)
		points[i+9] = plugins.PointWithTag("cpu.GuestNice", value.GuestNice, "core", key)

		i = i + 10
	}

	return points
}

func (c *CpuStats) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("CPU statistics")

	doc.AddTag("core", "The cpu core")

	doc.AddMeasurement("misc.Interrupts", "Number of interrupts per second", "/s")
	doc.AddMeasurement("misc.ContextSwitches", "Number of context switches per second", "/s")
	doc.AddMeasurement("misc.Forks", "Number of forks per second", "/s")
	doc.AddMeasurement("misc.RunningProcesses", "Currently running processe", "(n")
	doc.AddMeasurement("misc.BlockedProcesses", "Number of processes currently blocke", "(n")

	doc.AddMeasurement("cpu.User", "Time spend in user mode", "ticks/s")
	doc.AddMeasurement("cpu.Nice", "Time spend in user mode with low priority", "ticks/s")
	doc.AddMeasurement("cpu.System", "Time spend in kernel mode", "ticks/s")
	doc.AddMeasurement("cpu.Idle", "Time spend idle", "ticks/s")
	doc.AddMeasurement("cpu.IoWait", "Time spend waiting for IO", "ticks/s")
	doc.AddMeasurement("cpu.Irq", "Time spend processing interrupts", "ticks/s")
	doc.AddMeasurement("cpu.SoftIrq", "Time spend processing soft interrupts", "ticks/s")
	doc.AddMeasurement("cpu.Steal", "Time spend waiting for the *physical* CPU on a guest", "ticks/s")
	doc.AddMeasurement("cpu.Guest", "Time spend on running guests", "ticks/s")
	doc.AddMeasurement("cpu.GuestNice", "Time spend on running nice guests", "ticks/s")

	return doc
}

// Ensure compliance
var _ plugins.Agent = (*CpuStats)(nil)
