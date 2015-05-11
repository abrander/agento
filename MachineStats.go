package agento

import (
	"math"
	"os"
	"time"
)

type MachineStats struct {
	Hostname       string        `json:"h"`
	MemoryStats    *MemoryStats  `json:"m"`
	CpuStats       *CpuStats     `json:"c"`
	DiskStats      *DiskStats    `json:"d"`
	NetStats       *NetStats     `json:"n"`
	LoadStats      *LoadStats    `json:"l"`
	GatherDuration time.Duration `json:"g"`
}

func (m *MachineStats) Gather() {
	start := time.Now()

	m.Hostname, _ = os.Hostname()
	m.MemoryStats = GetMemoryStats()
	m.CpuStats = GetCpuStats()
	m.DiskStats = GetDiskStats()
	m.NetStats = GetNetStats()
	m.LoadStats = GetLoadStats()

	m.GatherDuration = time.Now().Sub(start)
}

func (s *MachineStats) GetMap() map[string]float64 {
	m := make(map[string]float64)

	s.MemoryStats.GetMap(m)
	s.CpuStats.GetMap(m)
	s.DiskStats.GetMap(m)
	s.NetStats.GetMap(m)
	s.LoadStats.GetMap(m)

	m["agento.gatherduration"] = Round(s.GatherDuration.Seconds()*1000.0, 1)

	return m
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
