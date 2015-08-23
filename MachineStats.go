package agento

import (
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/influxdb/influxdb/client"
)

type MachineStats struct {
	Hostname         string          `json:"h"`
	MemoryStats      *MemoryStats    `json:"m"`
	CpuStats         *CpuStats       `json:"c"`
	DiskStats        *DiskStats      `json:"d"`
	DiskUsageStats   *DiskUsageStats `json:"du"`
	NetStats         *NetStats       `json:"n"`
	LoadStats        *LoadStats      `json:"l"`
	SnmpStats        *SnmpStats      `json:"s"`
	CpuSpeed         *CpuSpeed       `json:"f"`
	SocketStats      *SocketStats    `json:"S"`
	AvailableEntropy int64           `json:"e"`
	GatherDuration   time.Duration   `json:"g"`
}

func (m *MachineStats) Gather() {
	start := time.Now()

	m.Hostname, _ = os.Hostname()
	m.MemoryStats = GetMemoryStats()
	m.CpuStats = GetCpuStats()
	m.DiskStats = GetDiskStats()
	m.DiskUsageStats = GetDiskUsageStats()
	m.NetStats = GetNetStats()
	m.LoadStats = GetLoadStats()
	m.SnmpStats = GetSnmpStats()
	m.CpuSpeed = GetCpuSpeed()
	m.SocketStats = GetSocketStats()
	m.AvailableEntropy, _ = getAvailableEntropy()

	m.GatherDuration = time.Now().Sub(start)
}

func SimplePoint(key string, value interface{}) client.Point {
	return client.Point{
		Measurement: key,
		Fields: map[string]interface{}{
			"value": value,
		},
	}
}

func PointWithTag(key string, value interface{}, tagKey string, tagValue string) client.Point {
	return client.Point{
		Tags: map[string]string{
			tagKey: tagValue,
		},
		Measurement: key,
		Fields: map[string]interface{}{
			"value": value,
		},
	}
}

func (s *MachineStats) GetPoints() []client.Point {
	points := make([]client.Point, 2, 300)

	points[0] = SimplePoint("misc.AvailableEntropy", s.AvailableEntropy)
	points[1] = SimplePoint("agento.GatherDuration", Round(s.GatherDuration.Seconds()*1000.0, 1))

	points = append(points, s.MemoryStats.GetPoints()...)
	points = append(points, s.CpuStats.GetPoints()...)
	points = append(points, s.DiskStats.GetPoints()...)
	points = append(points, s.DiskUsageStats.GetPoints()...)
	points = append(points, s.NetStats.GetPoints()...)
	points = append(points, s.LoadStats.GetPoints()...)
	points = append(points, s.SnmpStats.GetPoints()...)
	points = append(points, s.CpuSpeed.GetPoints()...)
	points = append(points, s.SocketStats.GetPoints()...)

	return points
}

func (s *MachineStats) GetDoc() map[string]string {
	m := make(map[string]string)

	s.MemoryStats.GetDoc(m)
	s.CpuStats.GetDoc(m)
	s.DiskStats.GetDoc(m)
	s.DiskUsageStats.GetDoc(m)
	s.NetStats.GetDoc(m)
	s.LoadStats.GetDoc(m)
	s.SnmpStats.GetDoc(m)
	s.CpuSpeed.GetDoc(m)
	s.SocketStats.GetDoc(m)
	m["misc.AvailableEntropy"] = "Available entropy in the kernel pool (b)"

	m["agento.GatherDuration"] = "Time used to gather metrics for Agento (ms)"

	return m
}

func getAvailableEntropy() (int64, error) {
	contents, err := ioutil.ReadFile("/proc/sys/kernel/random/entropy_avail")

	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(strings.TrimSpace(string(contents)), 10, 64)
}

func Round(value float64, places int) float64 {
	var round float64

	pow := math.Pow(10, float64(places))

	digit := pow * value
	_, div := math.Modf(digit)
	if div >= 0.5 {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}

	return round / pow
}
