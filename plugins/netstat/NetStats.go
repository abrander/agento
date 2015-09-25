package netstat

import (
	"bufio"
	"os"
	"path/filepath"
	//	"strconv"
	"strings"
	"time"

	"github.com/influxdb/influxdb/client"

	"github.com/abrander/agento"
	"github.com/abrander/agento/plugins"
)

func init() {
	plugins.Register("n", NewNetStats)
}

type NetStats struct {
	sampletime       time.Time `json:"-"`
	previousNetStats *NetStats
	Interfaces       map[string]*SingleNetStats `json:"ifs"`
}

func NewNetStats() plugins.Plugin {
	return new(NetStats)
}

func (n *NetStats) Gather() error {
	stat := NetStats{}
	stat.Interfaces = make(map[string]*SingleNetStats)

	path := filepath.Join("/proc/net/dev")
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
		if len(data) != 17 {
			continue
		}

		if strings.HasSuffix(data[0], ":") {
			s := SingleNetStats{}
			s.ReadArray(data)
			stat.Interfaces[strings.TrimSuffix(data[0], ":")] = &s
		}
	}

	*n = *stat.Sub(n.previousNetStats)
	n.previousNetStats = &stat

	return nil
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
		points[i+0] = agento.PointWithTag("net.RxBytes", value.RxBytes, "interface", key)
		points[i+1] = agento.PointWithTag("net.RxPackets", value.RxPackets, "interface", key)
		points[i+2] = agento.PointWithTag("net.RxErrors", value.RxErrors, "interface", key)
		points[i+3] = agento.PointWithTag("net.RxDropped", value.RxDropped, "interface", key)
		points[i+4] = agento.PointWithTag("net.RxFifo", value.RxFifo, "interface", key)
		points[i+5] = agento.PointWithTag("net.RxFrame", value.RxFrame, "interface", key)
		points[i+6] = agento.PointWithTag("net.RxCompressed", value.RxCompressed, "interface", key)
		points[i+7] = agento.PointWithTag("net.RxMulticast", value.RxMulticast, "interface", key)
		points[i+8] = agento.PointWithTag("net.TxBytes", value.TxBytes, "interface", key)
		points[i+9] = agento.PointWithTag("net.TxPackets", value.TxPackets, "interface", key)
		points[i+10] = agento.PointWithTag("net.TxErrors", value.TxErrors, "interface", key)
		points[i+11] = agento.PointWithTag("net.TxDropped", value.TxDropped, "interface", key)
		points[i+12] = agento.PointWithTag("net.TxFifo", value.TxFifo, "interface", key)
		points[i+13] = agento.PointWithTag("net.TxCollisions", value.TxCollisions, "interface", key)
		points[i+14] = agento.PointWithTag("net.TxCarrier", value.TxCarrier, "interface", key)
		points[i+15] = agento.PointWithTag("net.TxCompressed", value.TxCompressed, "interface", key)

		i = i + 16
	}

	return points
}

func (c *NetStats) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc()

	doc.AddMeasurement("net.RxBytes", "Bytes received", "b/s")
	doc.AddMeasurement("net.RxPackets", "Packets received", "packets/s")
	doc.AddMeasurement("net.RxErrors", "Receiver errors detected", "errors/s")
	doc.AddMeasurement("net.RxDropped", "Dropped packets", "packets/s")
	doc.AddMeasurement("net.RxFifo", "FIFO buffer overruns", "overruns/s")
	doc.AddMeasurement("net.RxFrame", "Framing errors", "errors/s")
	doc.AddMeasurement("net.RxCompressed", "Compressed frames received", "frames/s")
	doc.AddMeasurement("net.RxMulticast", "Multicast frames received", "frames/s")
	doc.AddMeasurement("net.TxBytes", "Bytes transmitted", "b/s")
	doc.AddMeasurement("net.TxPackets", "Packets transmitted", "packets/s")
	doc.AddMeasurement("net.TxErrors", "Transmission errors", "errors/s")
	doc.AddMeasurement("net.TxDropped", "Packets dropped", "packets/s")
	doc.AddMeasurement("net.TxFifo", "FIFO buffer overruns", "overruns/s")
	doc.AddMeasurement("net.TxCollisions", "Network collisions detected", "collisions/s")
	doc.AddMeasurement("net.TxCarrier", "Carrier losses", "losses/s")
	doc.AddMeasurement("net.TxCompressed", "Compressed frames transmitted", "frames/s")

	return doc
}
