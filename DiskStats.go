package agento

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/influxdb/influxdb/client"
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

func (d *DiskStats) GetPoints() []client.Point {
	points := make([]client.Point, len(d.Disks)*11)

	i := 0
	for key, value := range d.Disks {
		points[i+0] = PointWithTag("io.ReadsCompleted", value.ReadsCompleted, "device", key)
		points[i+1] = PointWithTag("io.ReadsMerged", value.ReadsMerged, "device", key)
		points[i+2] = PointWithTag("io.ReadSectors", value.ReadSectors, "device", key)
		points[i+3] = PointWithTag("io.ReadTime", value.ReadTime, "device", key)
		points[i+4] = PointWithTag("io.WritesCompleted", value.WritesCompleted, "device", key)
		points[i+5] = PointWithTag("io.WritesMerged", value.WritesMerged, "device", key)
		points[i+6] = PointWithTag("io.WriteSectors", value.WriteSectors, "device", key)
		points[i+7] = PointWithTag("io.WriteTime", value.WriteTime, "device", key)
		points[i+8] = PointWithTag("io.IoInProgress", value.IoInProgress, "device", key)
		points[i+9] = PointWithTag("io.IoTime", value.IoTime, "device", key)
		points[i+10] = PointWithTag("io.IoWeightedTime", value.IoWeightedTime, "device", key)

		i = i + 11
	}

	return points
}

func (c *DiskStats) GetDoc(m map[string]string) {
	m["io.ReadsCompleted"] = "Reads from device (reads/s)"
	m["io.ReadsMerged"] = "Reads merged (merges/s)"
	m["io.ReadSectors"] = "Sectors read (sectors/s)"
	m["io.ReadTime"] = "Milliseconds spend reading (ms/s)"
	m["io.WritesCompleted"] = "Writes to device (writes/s)"
	m["io.WritesMerged"] = "Writes merged (merges/s)"
	m["io.WriteSectors"] = "Sectors written (sectors/s"
	m["io.WriteTime"] = "Time spend writing (ms/s)"
	m["io.IoInProgress"] = "The current queue size of IO operations (n)"
	m["io.IoTime"] = "Time spend on IO (ms/s)"
	m["io.IoWeightedTime"] = "Time spend on IO times the IO queue. Please see https://www.kernel.org/doc/Documentation/iostats.txt (ms/s)"
}
