package socketstats

import (
	"path/filepath"
	"strings"

	"github.com/influxdata/influxdb/client/v2"

	"github.com/abrander/agento/configuration"
	"github.com/abrander/agento/plugins"
)

func init() {
	plugins.Register("sockets", NewSocketStats)
}

func NewSocketStats() plugins.Plugin {
	return new(SocketStats)
}

// https://www.kernel.org/doc/Documentation/networking/proc_net_tcp.txt

type SocketStats struct {
	Established int64 `json:"e"` // ESTABLISHED
	SynSent     int64 `json:"s"` // SYN_SENT1
	SynReceived int64 `json:"S"` // SYN_SENT2
	FinWait1    int64 `json:"f"` // FIN_WAIT1
	FinWait2    int64 `json:"F"` // FIN_WAIT2
	TimeWait    int64 `json:"t"` // TIME_WAIT
	Close       int64 `json:"c"` // CLOSE
	CloseWait   int64 `json:"C"` // CLOSE_WAIT
	LastAck     int64 `json:"a"` // LAST_ACK
	Listen      int64 `json:"l"` // LISTEN
	Closing     int64 `json:"o"` // CLOSING
	RootUser    int64 `json:"r"` // Sockets owned by root
}

func readFile(transport plugins.Transport, path string) []string {
	data, err := transport.ReadFile(path)
	if err != nil {
		return nil
	}

	lines := strings.Split(string(data), "\n")

	return lines[1:]
}

func (stats *SocketStats) Gather(transport plugins.Transport) error {
	// Reset before accumulating
	stats.Established = 0
	stats.SynSent = 0
	stats.SynReceived = 0
	stats.FinWait1 = 0
	stats.FinWait2 = 0
	stats.TimeWait = 0
	stats.Close = 0
	stats.CloseWait = 0
	stats.LastAck = 0
	stats.Listen = 0
	stats.Closing = 0
	stats.RootUser = 0

	tcpLines := readFile(transport, filepath.Join(configuration.ProcPath, "/net/tcp"))
	tcp6Lines := readFile(transport, filepath.Join(configuration.ProcPath, "/net/tcp6"))
	udpLines := readFile(transport, filepath.Join(configuration.ProcPath, "/net/udp"))
	udp6Lines := readFile(transport, filepath.Join(configuration.ProcPath, "/net/udp6"))

	sockets := append(tcpLines, tcp6Lines...)
	sockets = append(sockets, udpLines...)
	sockets = append(sockets, udp6Lines...)

	for _, line := range sockets {
		lineArray := strings.Fields(line)

		if len(lineArray) < 8 {
			continue
		}

		switch lineArray[3] {
		case "01": // ESTABLISHED
			stats.Established += 1
		case "02": // SYN_SENT1
			stats.SynSent += 1
		case "03": // SYN_SENT2
			stats.SynReceived += 1
		case "04": // FIN_WAIT1
			stats.FinWait1 += 1
		case "05": // FIN_WAIT2
			stats.FinWait2 += 1
		case "06": // TIME_WAIT
			stats.TimeWait += 1
		case "07": // CLOSE
			stats.Close += 1
		case "08": // CLOSE_WAIT
			stats.CloseWait += 1
		case "09": // LAST_ACK
			stats.LastAck += 1
		case "0A": // LISTEN
			stats.Listen += 1
		case "0B": // CLOSING
			stats.Closing += 1
		}

		if lineArray[7] == "0" {
			stats.RootUser += 1
		}
	}

	return nil
}

func (s *SocketStats) GetPoints() []*client.Point {
	points := make([]*client.Point, 12)

	points[0] = plugins.SimplePoint("sockets.Established", s.Established)
	points[1] = plugins.SimplePoint("sockets.SynSent", s.SynSent)
	points[2] = plugins.SimplePoint("sockets.SynReceived", s.SynReceived)
	points[3] = plugins.SimplePoint("sockets.FinWait1", s.FinWait1)
	points[4] = plugins.SimplePoint("sockets.FinWait2", s.FinWait2)
	points[5] = plugins.SimplePoint("sockets.TimeWait", s.TimeWait)
	points[6] = plugins.SimplePoint("sockets.Close", s.Close)
	points[7] = plugins.SimplePoint("sockets.CloseWait", s.CloseWait)
	points[8] = plugins.SimplePoint("sockets.LastAck", s.LastAck)
	points[9] = plugins.SimplePoint("sockets.Listen", s.Listen)
	points[10] = plugins.SimplePoint("sockets.Closing", s.Closing)
	points[11] = plugins.SimplePoint("sockets.RootUser", s.RootUser)

	return points
}

func (s *SocketStats) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("Socket usage")

	doc.AddMeasurement("sockets.Established", "Number of sockets in state ESTABLISHED", "n")
	doc.AddMeasurement("sockets.SynSent", "Number of sockets in state SYN_SENT1", "n")
	doc.AddMeasurement("sockets.SynReceived", "Number of sockets in state SYN_SENT2", "n")
	doc.AddMeasurement("sockets.FinWait1", "Number of sockets in state FIN_WAIT1", "n")
	doc.AddMeasurement("sockets.FinWait2", "Number of sockets in state FIN_WAIT2", "n")
	doc.AddMeasurement("sockets.TimeWait", "Number of sockets in state TIME_WAIT", "n")
	doc.AddMeasurement("sockets.Close", "Number of sockets in state CLOSE", "n")
	doc.AddMeasurement("sockets.CloseWait", "Number of sockets in state CLOSE_WAIT", "n")
	doc.AddMeasurement("sockets.LastAck", "Number of sockets in state LAST_ACK", "n")
	doc.AddMeasurement("sockets.Listen", "Number of sockets in state LISTEN", "n")
	doc.AddMeasurement("sockets.Closing", "Number of sockets in state CLOSING", "n")
	doc.AddMeasurement("sockets.RootUser", "Number of sockets owned by root", "n")

	return doc
}

// Ensure compliance
var _ plugins.Agent = (*SocketStats)(nil)
