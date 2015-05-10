package agento

import (
	"bufio"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type MachineStats struct {
	Hostname  string            `json:"h"`
	MemInfo   *map[string]int64 `json:"me"`
	CpuStats  *CpuStats         `json:"cp"`
	DiskStats *DiskStats        `json:"di"`
	NetStats  *NetStats         `json:"ne"`
	LoadStats *LoadStats        `json:"lo"`
}

func (m *MachineStats) Gather() {
	m.Hostname, _ = os.Hostname()

	m.MemInfo = getMemInfo()
	m.CpuStats = GetCpuStats()
	m.DiskStats = GetDiskStats()
	m.NetStats = GetNetStats()
	m.LoadStats = GetLoadStats()
}

func getMemInfo() *map[string]int64 {
	m := make(map[string]int64)

	path := filepath.Join("/proc/meminfo")
	file, err := os.Open(path)
	if err != nil {
		return &m
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()

		n := strings.Index(text, ":")
		if n == -1 {
			continue
		}

		key := text[:n]
		data := strings.Split(strings.Trim(text[(n+1):], " "), " ")
		if len(data) == 1 {
			value, err := strconv.ParseInt(data[0], 10, 64)
			if err != nil {
				continue
			}
			m[key] = value
		} else if len(data) == 2 {
			if data[1] == "kB" {
				value, err := strconv.ParseInt(data[0], 10, 64)
				if err != nil {
					continue
				}

				m[key] = value * 1024
			}
		}
	}

	return &m
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
