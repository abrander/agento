package agento

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var previousDiskStats *DiskStats

type DiskStats struct {
	sampletime time.Time                   `json:"-"`
	Disks      map[string]*SingleDiskStats `json:"disks"`
}

func GetDiskStats() *DiskStats {
	stat := DiskStats{}
	stat.Disks = make(map[string]*SingleDiskStats)

	path := filepath.Join("/proc/diskstats")
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
		if len(data) != 14 {
			continue
		}

		readsCompleted, _ := strconv.ParseInt(data[3], 10, 32)

		if readsCompleted > 0 {
			s := SingleDiskStats{}
			s.ReadArray(data)
			stat.Disks[data[2]] = &s
		}
	}

	ret := stat.Sub(previousDiskStats)
	previousDiskStats = &stat

	return ret
}

func (c *DiskStats) Sub(previousDiskStats *DiskStats) *DiskStats {
	if previousDiskStats == nil {
		return &DiskStats{}
	}

	diff := DiskStats{}
	diff.Disks = make(map[string]*SingleDiskStats)

	duration := float64(c.sampletime.Sub(previousDiskStats.sampletime)) / float64(time.Second)
	for key, value := range c.Disks {
		diff.Disks[key] = value.Sub(previousDiskStats.Disks[key], duration)
	}

	return &diff
}

func (c *DiskStats) GetMap(m map[string]interface{}) {
	if c == nil {
		return
	}

	if c.Disks == nil {
		return
	}

	for key, value := range c.Disks {
		m[key+".ReadsCompleted"] = value.ReadsCompleted
		m[key+".ReadsMerged"] = value.ReadsMerged
		m[key+".ReadSectors"] = value.ReadSectors
		m[key+".ReadTime"] = value.ReadTime
		m[key+".WritesCompleted"] = value.WritesCompleted
		m[key+".WritesMerged"] = value.WritesMerged
		m[key+".WriteSectors"] = value.WriteSectors
		m[key+".WriteTime"] = value.WriteTime
		m[key+".IoInProgress"] = value.IoInProgress
		m[key+".IoTime"] = value.IoTime
		m[key+".IoWeightedTime"] = value.IoWeightedTime
	}
}
