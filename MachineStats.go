package agento

import (
	"math"
	"os"
)

type MachineStats struct {
	Hostname    string       `json:"h"`
	MemoryStats *MemoryStats `json:"m"`
	CpuStats    *CpuStats    `json:"c"`
	DiskStats   *DiskStats   `json:"d"`
	NetStats    *NetStats    `json:"n"`
	LoadStats   *LoadStats   `json:"l"`
}

func (m *MachineStats) Gather() {
	m.Hostname, _ = os.Hostname()

	m.MemoryStats = GetMemoryStats()
	m.CpuStats = GetCpuStats()
	m.DiskStats = GetDiskStats()
	m.NetStats = GetNetStats()
	m.LoadStats = GetLoadStats()
}

func mapDelta(a *map[string]int64, b *map[string]int64) *map[string]int64 {
	m := make(map[string]int64)

	for key, value := range *a {
		if (*b)[key] != value {
			m[key] = (*b)[key] - value
		}
	}

	return &m
}

func unDelta(target *map[string]int64, diff *map[string]int64) error {
	for key, value := range *diff {
		(*target)[key] = (*target)[key] + value
	}

	return nil
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
