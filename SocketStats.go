package agento

import (
	"io/ioutil"
	"strings"

	"github.com/influxdb/influxdb/client"
)

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

func readFile(path string) []string {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}

	lines := strings.Split(string(data), "\n")

	return lines[1:]
}

func GetSocketStats() *SocketStats {
	stats := SocketStats{}

	tcpLines := readFile("/proc/net/tcp")
	tcp6Lines := readFile("/proc/net/tcp6")
	udpLines := readFile("/proc/net/udp")
	udp6Lines := readFile("/proc/net/udp6")

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

	return &stats
}

func (s *SocketStats) GetPoints() []client.Point {
	points := make([]client.Point, 12)

	points[0] = SimplePoint("sockets.Established", s.Established)
	points[1] = SimplePoint("sockets.SynSent", s.SynSent)
	points[2] = SimplePoint("sockets.SynReceived", s.SynReceived)
	points[3] = SimplePoint("sockets.FinWait1", s.FinWait1)
	points[4] = SimplePoint("sockets.FinWait2", s.FinWait2)
	points[5] = SimplePoint("sockets.TimeWait", s.TimeWait)
	points[6] = SimplePoint("sockets.Close", s.Close)
	points[7] = SimplePoint("sockets.CloseWait", s.CloseWait)
	points[8] = SimplePoint("sockets.LastAck", s.LastAck)
	points[9] = SimplePoint("sockets.Listen", s.Listen)
	points[10] = SimplePoint("sockets.Closing", s.Closing)
	points[11] = SimplePoint("sockets.RootUser", s.RootUser)

	return points
}

func (s *SocketStats) GetDoc(m map[string]string) {
	m["sockets.Established"] = "Number of sockets in state ESTABLISHED (n)"
	m["sockets.SynSent"] = "Number of sockets in state SYN_SENT1 (n)"
	m["sockets.SynReceived"] = "Number of sockets in state SYN_SENT2 (n)"
	m["sockets.FinWait1"] = "Number of sockets in state FIN_WAIT1 (n)"
	m["sockets.FinWait2"] = "Number of sockets in state FIN_WAIT2 (n)"
	m["sockets.TimeWait"] = "Number of sockets in state TIME_WAIT (n)"
	m["sockets.Close"] = "Number of sockets in state CLOSE (n)"
	m["sockets.CloseWait"] = "Number of sockets in state CLOSE_WAIT (n)"
	m["sockets.LastAck"] = "Number of sockets in state LAST_ACK (n)"
	m["sockets.Listen"] = "Number of sockets in state LISTEN (n)"
	m["sockets.Closing"] = "Number of sockets in state CLOSING (n)"
	m["sockets.RootUser"] = "Number of sockets owned by root (n)"
}
