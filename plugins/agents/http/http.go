package http

import (
	"net"
	"net/http"
	"time"

	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/timeseries"
)

func init() {
	plugins.Register("http", NewHttp)
}

func NewHttp() interface{} {
	return new(Http)
}

type (
	Http struct {
		Url             string `json:"url" description:"The URL to request"`
		Status          int
		Time            time.Duration
		ConnectDuration time.Duration
		RequestDuration time.Duration
	}

	instrumentTransport struct {
		rtp       http.RoundTripper
		dialer    func(network, addr string) (net.Conn, error)
		connStart time.Time
		connEnd   time.Time
		reqStart  time.Time
		reqEnd    time.Time
	}
)

func (h Http) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("Check response from http and https servers")

	doc.AddMeasurement("http", "Timing and status. Three values are provided: \"ConnectDuration\", \"RequestDuration\" and \"Status\"", "ms")
	doc.AddTag("url", "The requested URL")

	return doc
}

func newTransport(dialer func(network, addr string) (net.Conn, error)) *instrumentTransport {
	tr := &instrumentTransport{
		dialer: dialer,
	}

	tr.rtp = &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		Dial:                tr.dial,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableKeepAlives:   true,
	}

	return tr
}

func (tr *instrumentTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	tr.reqStart = time.Now()
	resp, err := tr.rtp.RoundTrip(r)
	tr.reqEnd = time.Now()

	return resp, err
}

func (tr *instrumentTransport) dial(network, addr string) (net.Conn, error) {
	tr.connStart = time.Now()
	cn, err := tr.dialer(network, addr)
	tr.connEnd = time.Now()

	return cn, err
}

func (tr *instrumentTransport) RequestDuration() time.Duration {
	return tr.Duration() - tr.ConnectDuration()
}

func (tr *instrumentTransport) ConnectDuration() time.Duration {
	return tr.connEnd.Sub(tr.connStart)
}

func (tr *instrumentTransport) Duration() time.Duration {
	return tr.reqEnd.Sub(tr.reqStart)
}

func (h *Http) Gather(transport plugins.Transport) error {
	tr := newTransport(transport.Dial)
	client := &http.Client{Transport: tr}
	resp, err := client.Get(h.Url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	h.Status = resp.StatusCode
	h.Time = tr.Duration()
	h.ConnectDuration = tr.ConnectDuration()
	h.RequestDuration = tr.RequestDuration()

	return nil
}

func (h Http) GetPoints() []*timeseries.Point {
	p := make([]*timeseries.Point, 1)

	p[0] = timeseries.NewPoint(
		"http",
		map[string]string{
			"url": h.Url,
		},
		map[string]interface{}{
			"ConnectDuration": h.ConnectDuration.Seconds() * 1000.0,
			"RequestDuration": h.RequestDuration.Seconds() * 1000.0,
			"Status":          h.Status,
		},
	)

	return p
}

// Ensure compliance
var _ plugins.Agent = (*Http)(nil)
