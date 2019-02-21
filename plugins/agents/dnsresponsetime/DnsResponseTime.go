package dnsresponsetime

import (
	"strings"

	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/timeseries"
	"github.com/miekg/dns"
)

type Data struct {
	Loss   int    `json:"loss"`
	Time   int64  `json:"time"`
	Domain string `json:"domain"`
	Server string `json:"server"`
}

type DnsResponseTime struct {
	Data []Data `json:"data"`

	Domains string `toml:"domain" json:"domain" description:"The domain(s) name to query (multiple can be separated by comma)"`
	Servers string `toml:"server" json:"server" description:"The server(s) to query (multiple can be separated by comma)"`
}

func init() {
	plugins.Register("dnsresponsetime", NewDnsResponseTime)
}

func NewDnsResponseTime() interface{} {
	return new(DnsResponseTime)
}

func (d *DnsResponseTime) Gather(transport plugins.Transport) error {
	domains := strings.Split(d.Domains, ",")
	servers := strings.Split(d.Servers, ",")

	for _, domain := range domains {
		for _, server := range servers {
			data := Data{}

			c := dns.Client{}
			m := dns.Msg{}
			m.SetQuestion(domain+".", dns.TypeA)
			_, t, err := c.Exchange(&m, server+":53")
			if err != nil {
				data.Time = 0
				data.Loss = 100
			} else {
				data.Time = t.Nanoseconds()
				data.Loss = 0
			}

			data.Domain = domain
			data.Server = server

			d.Data = append(d.Data, data)
		}
	}

	return nil
}

func (d *DnsResponseTime) GetPoints() []*timeseries.Point {
	points := make([]*timeseries.Point, len(d.Data))
	for i, data := range d.Data {
		tags := map[string]string{
			"domain": data.Domain,
			"server": data.Server,
		}

		values := map[string]interface{}{
			"loss":        data.Loss,
			"nanoseconds": data.Time,
		}

		points[i] = plugins.PointValuesWithTags("dnsresponsetime", values, tags)
	}

	return points
}

func (m *DnsResponseTime) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("DNS Response Time")

	doc.AddTag("domain", "The domain(s) name to query (multiple can be separated by comma)")
	doc.AddTag("server", "The server(s) to query (multiple can be separated by comma)")
	doc.AddMeasurement("dnsresponsetime.time", "Time it took to query the server in nanoseconds", "n")
	doc.AddMeasurement("dnsresponsetime.loss", "Amount of loss (timeouts) in percent", "n")

	return doc
}

// Ensure compliance
var _ plugins.Agent = (*DnsResponseTime)(nil)
