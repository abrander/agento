package agento

import (
	"syscall"
)

type SingleDiskUsageStats struct {
	Used      uint64 `json:"u"`
	Reserved  uint64 `json:"r"`
	Free      uint64 `json:"f"`
	UsedNodes uint64 `json:"un"`
	FreeNodes uint64 `json:"fn"`
}

func ReadSingleDiskUsageStats(path string) *SingleDiskUsageStats {
	var stats SingleDiskUsageStats
	var stat syscall.Statfs_t

	err := syscall.Statfs(path, &stat)
	if err != nil {
		return nil
	}

	bSize := uint64(stat.Bsize)
	stats.Used = bSize * (stat.Blocks - stat.Bfree)
	stats.Reserved = bSize * (stat.Bfree - stat.Bavail)
	stats.Free = bSize * stat.Bavail

	stats.UsedNodes = stat.Files - stat.Ffree
	stats.FreeNodes = stat.Ffree

	return &stats
}

/*





Type    int64	61267
Bsize   int64	4096
Blocks  uint64	37910569
Bfree   uint64	33643847
Bavail  uint64	31712327
Files   uint64	9641984
Ffree   uint64	9128527

Fsid    Fsid
Namelen int64
Frsize  int64
Flags   int64
Spare   [4]int64
*/
