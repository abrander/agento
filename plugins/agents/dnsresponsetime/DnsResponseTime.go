package dnsresponsetime

import (
	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/timeseries"
	"github.com/miekg/dns"
)

type Data struct {
	Loss int   `json:"loss"`
	Time int64 `json:"time"`
}

type DnsResponseTime struct {
	Data []Data `json:"data"`

	Domain string `toml:"domain" json:"domain" description:"Domain(s) to query"`
	Server string `toml:"server" json:"server" description:"The server(s) to query"`
}

func init() {
	plugins.Register("dnsresponsetime", NewDnsResponseTime)
}

func NewDnsResponseTime() interface{} {
	return new(DnsResponseTime)
}

func (d *DnsResponseTime) Gather(transport plugins.Transport) error {
	data := Data{}

	c := dns.Client{}
	m := dns.Msg{}
	m.SetQuestion(d.Domain+".", dns.TypeA)
	_, t, err := c.Exchange(&m, d.Server+":53")
	if err != nil {
		data.Time = 0
		data.Loss = 100
	} else {
		data.Time = t.Nanoseconds()
		data.Loss = 0
	}

	d.Data = append(d.Data, data)

	return nil
}

func (d *DnsResponseTime) GetPoints() []*timeseries.Point {
	points := make([]*timeseries.Point, len(d.Data))
	for i, data := range d.Data {

		tags := map[string]string{
			"domain": d.Domain,
			"server": d.Server,
		}

		values := map[string]interface{}{
			"loss": data.Loss,
			"time": data.Time,
		}

		points[i] = plugins.PointValuesWithTags("dnsresponsetime", values, tags)
	}
	return points
}

func (m *DnsResponseTime) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("DNS Response Time")

	doc.AddTag("domain", "The domain name to query")
	doc.AddTag("server", "The server to query")
	doc.AddMeasurement("dnsresponsetime.time", "Time it took to query the server in nanoseconds", "n")
	doc.AddMeasurement("dnsresponsetime.loss", "Amount of loss (timeouts) in percent", "n")

	return doc
}

// Ensure compliance
var _ plugins.Agent = (*DnsResponseTime)(nil)
