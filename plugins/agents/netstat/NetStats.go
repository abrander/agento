package netstat

import (
	"bufio"
	"path/filepath"
	//	"strconv"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"

	"github.com/abrander/agento/configuration"
	"github.com/abrander/agento/plugins"
)

func init() {
	plugins.Register("netio", NewNetStats)
}

type NetStats struct {
	sampletime       time.Time `json:"-"`
	previousNetStats *NetStats
	Interfaces       map[string]*SingleNetStats `json:"ifs"`
}

func NewNetStats() plugins.Plugin {
	return new(NetStats)
}

func (stat *NetStats) Gather(transport plugins.Transport) error {
	stat.Interfaces = make(map[string]*SingleNetStats)

	path := filepath.Join(configuration.ProcPath, "/net/dev")
	file, err := transport.Open(path)
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

	return nil
}

func (n *NetStats) GetPoints() []*client.Point {
	points := make([]*client.Point, len(n.Interfaces)*16)

	i := 0
	for key, value := range n.Interfaces {
		points[i+0] = plugins.PointWithTag("net.RxBytes", value.RxBytes, "interface", key)
		points[i+1] = plugins.PointWithTag("net.RxPackets", value.RxPackets, "interface", key)
		points[i+2] = plugins.PointWithTag("net.RxErrors", value.RxErrors, "interface", key)
		points[i+3] = plugins.PointWithTag("net.RxDropped", value.RxDropped, "interface", key)
		points[i+4] = plugins.PointWithTag("net.RxFifo", value.RxFifo, "interface", key)
		points[i+5] = plugins.PointWithTag("net.RxFrame", value.RxFrame, "interface", key)
		points[i+6] = plugins.PointWithTag("net.RxCompressed", value.RxCompressed, "interface", key)
		points[i+7] = plugins.PointWithTag("net.RxMulticast", value.RxMulticast, "interface", key)
		points[i+8] = plugins.PointWithTag("net.TxBytes", value.TxBytes, "interface", key)
		points[i+9] = plugins.PointWithTag("net.TxPackets", value.TxPackets, "interface", key)
		points[i+10] = plugins.PointWithTag("net.TxErrors", value.TxErrors, "interface", key)
		points[i+11] = plugins.PointWithTag("net.TxDropped", value.TxDropped, "interface", key)
		points[i+12] = plugins.PointWithTag("net.TxFifo", value.TxFifo, "interface", key)
		points[i+13] = plugins.PointWithTag("net.TxCollisions", value.TxCollisions, "interface", key)
		points[i+14] = plugins.PointWithTag("net.TxCarrier", value.TxCarrier, "interface", key)
		points[i+15] = plugins.PointWithTag("net.TxCompressed", value.TxCompressed, "interface", key)

		i = i + 16
	}

	return points
}

func (c *NetStats) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("Network")

	doc.AddTag("interface", "The network interface")

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

// Ensure compliance
var _ plugins.Agent = (*NetStats)(nil)
