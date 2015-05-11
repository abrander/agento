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
