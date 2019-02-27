package ping

import (
	"strings"
	"time"

	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/timeseries"
	gping "github.com/gansoi/gansoi/plugins/agents/ping"
)

type Data struct {
	Loss    int    `json:"loss"`
	Sent    int    `json:"sent"`
	Replies int    `json:"replies"`
	TimeAvg int64  `json:"timeavg"`
	TimeMin int64  `json:"timemin"`
	TimeMax int64  `json:"timemax"`
	Ip      string `json:"server"`
}

type Ping struct {
	Data   []Data `json:"data"`
	pinger *gping.ICMPService

	Ip    string `toml:"ip" json:"ip" description:"The ip(s) to ping (multiple can be separated by comma)"`
	Count int    `toml:"count" json:"count" description:"Number of packages to send"`
}

func init() {
	plugins.Register("ping", NewPing)
}

func NewPing() interface{} {
	return new(Ping)
}

func (p *Ping) Gather(transport plugins.Transport) error {
	// Create and start the ICMP pinger service
	if p.pinger == nil {
		p.pinger = gping.NewICMPService()
		p.pinger.Start()
	}

	// Default to "one ping only" if nothing else is specified in config
	count := 1
	if p.Count > 0 {
		count = p.Count
	}

	ips := strings.Split(p.Ip, ",")
	for _, ip := range ips {
		data := Data{}

		summary, err := p.pinger.Ping(ip, count, time.Second)

		if err != nil {
			return err
		}

		data.Loss = 100 - (100 * summary.Replies / summary.Sent)
		data.Sent = summary.Sent
		data.Replies = summary.Replies
		data.TimeAvg = summary.Average.Nanoseconds()
		data.TimeMin = summary.Min.Nanoseconds()
		data.TimeMax = summary.Max.Nanoseconds()

		data.Ip = ip

		p.Data = append(p.Data, data)
	}
	return nil
}

func (p *Ping) GetPoints() []*timeseries.Point {
	points := make([]*timeseries.Point, len(p.Data))
	for i, data := range p.Data {
		tags := map[string]string{
			"ip": data.Ip,
		}

		values := map[string]interface{}{
			"loss":    data.Loss,
			"sent":    data.Sent,
			"replies": data.Replies,
			"timeAvg": data.TimeAvg,
			"timeMin": data.TimeMin,
			"timeMax": data.TimeMax,
		}

		points[i] = plugins.PointValuesWithTags("ping", values, tags)
	}

	return points
}

func (m *Ping) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("Ping Time")

	doc.AddTag("ip", "The ip(s) to ping (multiple can be separated by comma)")
	doc.AddTag("count", "Number of packages to send")
	doc.AddMeasurement("data.Loss", "Packetloss in percent", "n")
	doc.AddMeasurement("data.Sent", "Number of pings sent", "n")
	doc.AddMeasurement("data.Replies", "Number of replies", "n")
	doc.AddMeasurement("data.TimeAvg", "Avg time for all replies", "n")
	doc.AddMeasurement("data.TimeMin", "Min time for a reply", "n")
	doc.AddMeasurement("data.TimeMax", "Max time for a reply", "n")

	return doc
}

// Ensure compliance
var _ plugins.Agent = (*Ping)(nil)
