package agento

import ()

type DiskUsageStats struct {
	Disks map[string]*SingleDiskUsageStats `json:"disks"`
}

func GetDiskUsageStats() *DiskUsageStats {
	stat := DiskUsageStats{}
	stat.Disks = make(map[string]*SingleDiskUsageStats)

	stat.Disks["/"] = ReadSingleDiskUsageStats("/")

	return &stat
}

func (c *DiskUsageStats) GetMap(m map[string]float64) {
	if c == nil {
		return
	}

	if c.Disks == nil {
		return
	}

	for key, value := range c.Disks {
		m["du."+key+".Used"] = float64(value.Used)
		m["du."+key+".Reserved"] = float64(value.Reserved)
		m["du."+key+".Free"] = float64(value.Free)
		m["du."+key+".UsedNodes"] = float64(value.UsedNodes)
		m["du."+key+".FreeNodes"] = float64(value.FreeNodes)
	}
}
