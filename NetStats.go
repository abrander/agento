package agento

import (
	"bufio"
	"os"
	"path/filepath"
	//	"strconv"
	"strings"
	"time"

	"github.com/influxdb/influxdb/client"
)

var previousNetStats *NetStats

type NetStats struct {
	sampletime time.Time                  `json:"-"`
	Interfaces map[string]*SingleNetStats `json:"ifs"`
}

func GetNetStats() *NetStats {
	stat := NetStats{}
	stat.Interfaces = make(map[string]*SingleNetStats)

	path := filepath.Join("/proc/net/dev")
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
		if len(data) != 17 {
			continue
		}

		if strings.HasSuffix(data[0], ":") {
			s := SingleNetStats{}
			s.ReadArray(data)
			stat.Interfaces[strings.TrimSuffix(data[0], ":")] = &s
		}
	}

	ret := stat.Sub(previousNetStats)
	previousNetStats = &stat

	return ret
}

func (c *NetStats) Sub(previousNetStats *NetStats) *NetStats {
	if previousNetStats == nil {
		return &NetStats{}
	}

	diff := NetStats{}
	diff.Interfaces = make(map[string]*SingleNetStats)

	duration := float64(c.sampletime.Sub(previousNetStats.sampletime)) / float64(time.Second)
	for key, value := range c.Interfaces {
		diff.Interfaces[key] = value.Sub(previousNetStats.Interfaces[key], duration)
	}

	return &diff
}

func (n *NetStats) GetPoints() []client.Point {
	points := make([]client.Point, len(n.Interfaces)*16)

	i := 0
	for key, value := range n.Interfaces {
		points[i+0] = PointWithTag("net.RxBytes", value.RxBytes, "interface", key)
		points[i+1] = PointWithTag("net.RxPackets", value.RxPackets, "interface", key)
		points[i+2] = PointWithTag("net.RxErrors", value.RxErrors, "interface", key)
		points[i+3] = PointWithTag("net.RxDropped", value.RxDropped, "interface", key)
		points[i+4] = PointWithTag("net.RxFifo", value.RxFifo, "interface", key)
		points[i+5] = PointWithTag("net.RxFrame", value.RxFrame, "interface", key)
		points[i+6] = PointWithTag("net.RxCompressed", value.RxCompressed, "interface", key)
		points[i+7] = PointWithTag("net.RxMulticast", value.RxMulticast, "interface", key)
		points[i+8] = PointWithTag("net.TxBytes", value.TxBytes, "interface", key)
		points[i+9] = PointWithTag("net.TxPackets", value.TxPackets, "interface", key)
		points[i+10] = PointWithTag("net.TxErrors", value.TxErrors, "interface", key)
		points[i+11] = PointWithTag("net.TxDropped", value.TxDropped, "interface", key)
		points[i+12] = PointWithTag("net.TxFifo", value.TxFifo, "interface", key)
		points[i+13] = PointWithTag("net.TxCollisions", value.TxCollisions, "interface", key)
		points[i+14] = PointWithTag("net.TxCarrier", value.TxCarrier, "interface", key)
		points[i+15] = PointWithTag("net.TxCompressed", value.TxCompressed, "interface", key)

		i = i + 16
	}

	return points
}

func (c *NetStats) GetDoc(m map[string]string) {
	m["net.RxBytes"] = "Bytes received (b/s)"
	m["net.RxPackets"] = "Packets received (packets/s"
	m["net.RxErrors"] = "Receiver errors detected (errors/s)"
	m["net.RxDropped"] = "Dropped packets (packets/s)"
	m["net.RxFifo"] = "FIFO buffer overruns (overruns/s)"
	m["net.RxFrame"] = "Framing errors (errors/s)"
	m["net.RxCompressed"] = "Compressed frames received (frames/s)"
	m["net.RxMulticast"] = "Multicast frames received (frames/s)"
	m["net.TxBytes"] = "Bytes transmitted (b/s)"
	m["net.TxPackets"] = "Packets transmitted (packets/s)"
	m["net.TxErrors"] = "Transmission errors (errors/s)"
	m["net.TxDropped"] = "Packets dropped (packets/s)"
	m["net.TxFifo"] = "FIFO buffer overruns (overruns/s)"
	m["net.TxCollisions"] = "Network collisions detected (collisions/s)"
	m["net.TxCarrier"] = "Carrier losses (losses/s)"
	m["net.TxCompressed"] = "Compressed frames transmitted (frames/s)"
}
