package agento

import (
	"github.com/influxdb/influxdb/client"
)

type DiskUsageStats struct {
	Disks map[string]*SingleDiskUsageStats `json:"disks"`
}

func GetDiskUsageStats() *DiskUsageStats {
	stat := DiskUsageStats{}
	stat.Disks = make(map[string]*SingleDiskUsageStats)

	stat.Disks["/"] = ReadSingleDiskUsageStats("/")

	return &stat
}

func (d *DiskUsageStats) GetPoints() []client.Point {
	points := make([]client.Point, len(d.Disks)*5)

	i := 0
	for key, value := range d.Disks {
		points[i+0] = PointWithTag("du.Used", value.Used, "mountpoint", key)
		points[i+1] = PointWithTag("du.Reserved", value.Reserved, "mountpoint", key)
		points[i+2] = PointWithTag("du.Free", value.Free, "mountpoint", key)
		points[i+3] = PointWithTag("du.UsedNodes", value.UsedNodes, "mountpoint", key)
		points[i+4] = PointWithTag("du.FreeNodes", value.FreeNodes, "mountpoint", key)

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
