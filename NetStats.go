package agento

import (
	"bufio"
	"os"
	"path/filepath"
	//	"strconv"
	"strings"
	"time"
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

func (c *NetStats) GetMap() *map[string]float64 {
	m := make(map[string]float64)

	if c == nil {
		return &m
	}

	if c.Interfaces == nil {
		return &m
	}

	for key, value := range c.Interfaces {
		m[key+".RxBytes"] = value.RxBytes
		m[key+".RxPackets"] = value.RxPackets
		m[key+".RxErrors"] = value.RxErrors
		m[key+".RxDropped"] = value.RxDropped
		m[key+".RxFifo"] = value.RxFifo
		m[key+".RxFrame"] = value.RxFrame
		m[key+".RxCompressed"] = value.RxCompressed
		m[key+".RxMulticast"] = value.RxMulticast
		m[key+".TxBytes"] = value.TxBytes
		m[key+".TxPackets"] = value.TxPackets
		m[key+".TxErrors"] = value.TxErrors
		m[key+".TxDropped"] = value.TxDropped
		m[key+".TxFifo"] = value.TxFifo
		m[key+".TxCollisions"] = value.TxCollisions
		m[key+".TxCarrier"] = value.TxCarrier
		m[key+".TxCompressed"] = value.TxCompressed
	}

	return &m
}
