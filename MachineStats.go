package agento

import (
	"bufio"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type MachineStats struct {
	FrameType    string            `json:"frame"`
	Hostname     string            `json:"h"`
	MemInfo      *map[string]int64 `json:"me"`
	VmStat       *map[string]int64 `json:"vm"`
	CpuStats     *CpuStats         `json:"cp"`
	lastMemInfo  *map[string]int64 `json:"-"`
	lastVmStat   *map[string]int64 `json:"-"`
	lastCpuStats *CpuStats         `json:"-"`
}

func (m *MachineStats) Gather() {
	m.FrameType = "full"
	m.Hostname, _ = os.Hostname()

	m.lastMemInfo = m.MemInfo
	m.MemInfo = getMemInfo()

	m.lastVmStat = m.VmStat
	m.VmStat = getVmStat()

	m.CpuStats = GetCpuStats()
}

func (m *MachineStats) GetJson(delta bool) ([]byte, error) {
	if delta {
		if m.lastMemInfo == nil || m.lastVmStat == nil {
			return nil, errors.New("No previous data to diff")
		}

		// Create new MachineStats
		machineStats := MachineStats{}
		machineStats.FrameType = "delta"
		machineStats.Hostname = m.Hostname

		machineStats.MemInfo = mapDelta(m.lastMemInfo, m.MemInfo)
		machineStats.VmStat = mapDelta(m.lastVmStat, m.VmStat)
		machineStats.CpuStats = m.CpuStats

		return json.Marshal(machineStats)
	} else {
		return json.Marshal(m)
	}
}

func (m *MachineStats) ReadJson(jsonBlob []byte) error {
	newStats := MachineStats{}

	err := json.Unmarshal(jsonBlob, &newStats)
	if err != nil {
		return err
	}

	m.FrameType = "full"
	m.Hostname = newStats.Hostname

	if newStats.FrameType == "full" {
		m.MemInfo = newStats.MemInfo
		m.VmStat = newStats.VmStat
		m.CpuStats = newStats.CpuStats
	} else if newStats.FrameType == "delta" {
		if m.MemInfo == nil || m.VmStat == nil {
			return errors.New("No full frames received (yet)")
		}

		unDelta(m.MemInfo, newStats.MemInfo)
		unDelta(m.VmStat, newStats.VmStat)
		m.CpuStats = newStats.CpuStats
	} else {
		return errors.New("Unknown frametype received")
	}

	return nil
}

func getVmStat() *map[string]int64 {
	m := make(map[string]int64)

	path := filepath.Join("/proc/vmstat")
	file, err := os.Open(path)
	if err != nil {
		return &m
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		data := strings.Split(text, " ")
		if len(data) == 2 {
			key := data[0]
			value, err := strconv.ParseInt(data[1], 10, 64)
			if err != nil {
				continue
			}
			m[key] = value
		}
	}

	return &m
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
