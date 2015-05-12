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

func (c *DiskUsageStats) GetMap(m map[string]interface{}) {
	if c == nil {
		return
	}

	if c.Disks == nil {
		return
	}

	for key, value := range c.Disks {
		m["du."+key+".Used"] = value.Used
		m["du."+key+".Reserved"] = value.Reserved
		m["du."+key+".Free"] = value.Free
		m["du."+key+".UsedNodes"] = value.UsedNodes
		m["du."+key+".FreeNodes"] = value.FreeNodes
	}
}
