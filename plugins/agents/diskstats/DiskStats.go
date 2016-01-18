package diskstats

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
	plugins.Register("diskio", NewDiskStats)
}

func NewDiskStats() plugins.Plugin {
	return new(DiskStats)
}

type DiskStats struct {
	Disks map[string]*SingleDiskStats `json:"disks"`
}

func (stat *DiskStats) Gather(transport plugins.Transport) error {
	stat.Disks = make(map[string]*SingleDiskStats)

	path := filepath.Join(configuration.ProcPath, "/diskstats")
	file, err := transport.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

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

	return nil
}

func (d *DiskStats) GetPoints() []*client.Point {
	points := make([]*client.Point, len(d.Disks)*11)

	i := 0
	for key, value := range d.Disks {
		points[i+0] = plugins.PointWithTag("io.ReadsCompleted", value.ReadsCompleted, "device", key)
		points[i+1] = plugins.PointWithTag("io.ReadsMerged", value.ReadsMerged, "device", key)
		points[i+2] = plugins.PointWithTag("io.ReadSectors", value.ReadSectors, "device", key)
		points[i+3] = plugins.PointWithTag("io.ReadTime", value.ReadTime, "device", key)
		points[i+4] = plugins.PointWithTag("io.WritesCompleted", value.WritesCompleted, "device", key)
		points[i+5] = plugins.PointWithTag("io.WritesMerged", value.WritesMerged, "device", key)
		points[i+6] = plugins.PointWithTag("io.WriteSectors", value.WriteSectors, "device", key)
		points[i+7] = plugins.PointWithTag("io.WriteTime", value.WriteTime, "device", key)
		points[i+8] = plugins.PointWithTag("io.IoInProgress", value.IoInProgress, "device", key)
		points[i+9] = plugins.PointWithTag("io.IoTime", value.IoTime, "device", key)
		points[i+10] = plugins.PointWithTag("io.IoWeightedTime", value.IoWeightedTime, "device", key)

		i = i + 11
	}

	return points
}

func (c *DiskStats) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("IO")

	doc.AddTag("device", "The block device")

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

// Ensure compliance
var _ plugins.Agent = (*DiskStats)(nil)
