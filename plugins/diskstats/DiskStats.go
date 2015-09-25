package diskstats

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/influxdb/influxdb/client"

	"github.com/abrander/agento"
	"github.com/abrander/agento/plugins"
)

func init() {
	plugins.Register("d", NewDiskStats)
}

func NewDiskStats() plugins.Plugin {
	return new(DiskStats)
}

type DiskStats struct {
	sampletime        time.Time `json:"-"`
	previousDiskStats *DiskStats
	Disks             map[string]*SingleDiskStats `json:"disks"`
}

func (d *DiskStats) Gather() error {
	stat := DiskStats{}
	stat.Disks = make(map[string]*SingleDiskStats)

	path := filepath.Join("/proc/diskstats")
	file, err := os.Open(path)
	if err != nil {
		return err
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

	*d = *stat.Sub(d.previousDiskStats)
	d.previousDiskStats = &stat

	return nil
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
		points[i+0] = agento.PointWithTag("io.ReadsCompleted", value.ReadsCompleted, "device", key)
		points[i+1] = agento.PointWithTag("io.ReadsMerged", value.ReadsMerged, "device", key)
		points[i+2] = agento.PointWithTag("io.ReadSectors", value.ReadSectors, "device", key)
		points[i+3] = agento.PointWithTag("io.ReadTime", value.ReadTime, "device", key)
		points[i+4] = agento.PointWithTag("io.WritesCompleted", value.WritesCompleted, "device", key)
		points[i+5] = agento.PointWithTag("io.WritesMerged", value.WritesMerged, "device", key)
		points[i+6] = agento.PointWithTag("io.WriteSectors", value.WriteSectors, "device", key)
		points[i+7] = agento.PointWithTag("io.WriteTime", value.WriteTime, "device", key)
		points[i+8] = agento.PointWithTag("io.IoInProgress", value.IoInProgress, "device", key)
		points[i+9] = agento.PointWithTag("io.IoTime", value.IoTime, "device", key)
		points[i+10] = agento.PointWithTag("io.IoWeightedTime", value.IoWeightedTime, "device", key)

		i = i + 11
	}

	return points
}

func (c *DiskStats) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc()

	doc.AddMeasurement("io.ReadsCompleted", "Reads from device", "reads/s")
	doc.AddMeasurement("io.ReadsMerged", "Reads merged", "merges/s")
	doc.AddMeasurement("io.ReadSectors", "Sectors read", "sectors/s")
	doc.AddMeasurement("io.ReadTime", "Milliseconds spend reading", "ms/s")
	doc.AddMeasurement("io.WritesCompleted", "Writes to device", "writes/s")
	doc.AddMeasurement("io.WritesMerged", "Writes merged", "merges/s")
	doc.AddMeasurement("io.WriteSectors", "Sectors written", "sectors/")
	doc.AddMeasurement("io.WriteTime", "Time spend writing", "ms/s")
	doc.AddMeasurement("io.IoInProgress", "The current queue size of IO operation", "(n")
	doc.AddMeasurement("io.IoTime", "Time spend on IO", "ms/s")
	doc.AddMeasurement("io.IoWeightedTime", "Time spend on IO times the IO queue. Please see https://www.kernel.org/doc/Documentation/iostats.txt", "ms/s")

	return doc
}
