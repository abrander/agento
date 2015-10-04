package diskusage

import (
	"syscall"
)

type SingleDiskUsageStats struct {
	Used      int64 `json:"u"`
	Reserved  int64 `json:"r"`
	Free      int64 `json:"f"`
	UsedNodes int64 `json:"un"`
	FreeNodes int64 `json:"fn"`
}

func ReadSingleDiskUsageStats(path string) *SingleDiskUsageStats {
	var stats SingleDiskUsageStats
	var stat syscall.Statfs_t

	err := syscall.Statfs(path, &stat)
	if err != nil {
		return nil
	}

	bSize := int64(stat.Bsize)
	stats.Used = bSize * int64(stat.Blocks-stat.Bfree)
	stats.Reserved = bSize * int64(stat.Bfree-stat.Bavail)
	stats.Free = bSize * int64(stat.Bavail)

	stats.UsedNodes = int64(stat.Files - stat.Ffree)
	stats.FreeNodes = int64(stat.Ffree)

	return &stats
}
