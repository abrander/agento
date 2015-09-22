package agento

import (
	"github.com/influxdb/influxdb/client"

	"github.com/abrander/agento"
	"github.com/abrander/agento/plugins"
)

func init() {
	plugins.Register("du", NewDiskUsageStats)
}

func NewDiskUsageStats() plugins.Plugin {
	return new(DiskUsageStats)
}

type DiskUsageStats struct {
	Disks map[string]*SingleDiskUsageStats `json:"disks"`
}

func (du *DiskUsageStats) Gather() error {
	du.Disks = make(map[string]*SingleDiskUsageStats)

	// FIXME: Add dynamic disks
	du.Disks["/"] = ReadSingleDiskUsageStats("/")

	return nil
}

func (d *DiskUsageStats) GetPoints() []client.Point {
	points := make([]client.Point, len(d.Disks)*5)

	i := 0
	for key, value := range d.Disks {
		points[i+0] = agento.PointWithTag("du.Used", value.Used, "mountpoint", key)
		points[i+1] = agento.PointWithTag("du.Reserved", value.Reserved, "mountpoint", key)
		points[i+2] = agento.PointWithTag("du.Free", value.Free, "mountpoint", key)
		points[i+3] = agento.PointWithTag("du.UsedNodes", value.UsedNodes, "mountpoint", key)
		points[i+4] = agento.PointWithTag("du.FreeNodes", value.FreeNodes, "mountpoint", key)

		i = i + 5
	}

	return points
}

func (c *DiskUsageStats) GetDoc(m map[string]string) {
	m["du.Used"] = "Used space (b)"
	m["du.Reserved"] = "Space reserved for uid 0 (b)"
	m["du.Free"] = "Free space (b)"
	m["du.UsedNodes"] = "Used inodes (n)"
	m["du.FreeNodes"] = "Free inodes (n)"
}
