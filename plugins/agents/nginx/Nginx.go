package nginx

import (
	"fmt"

	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/timeseries"
)

func init() {
	plugins.Register("nginx", newNginx)
}

// Nginx will retrieve stub status.
type Nginx struct {
	URL string `toml:"url" description:"Nginx status URL"`

	ActiveConnections int
	Accepts           int
	Handled           int
	Requests          int
	Reading           int
	Writing           int
	Waiting           int
}

const stubFormat = `Active connections: %d
server accepts handled requests
 %d %d %d
Reading: %d Writing: %d Waiting: %d
`

func newNginx() interface{} {
	return new(Nginx)
}

// Gather will measure how many bytes can be read from /dev/null.
func (n *Nginx) Gather(transport plugins.Transport) error {
	client := plugins.HTTPClient(transport)
	resp, err := client.Get(n.URL)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	fmt.Fscanf(resp.Body, stubFormat,
		&n.ActiveConnections,
		&n.Accepts,
		&n.Handled,
		&n.Requests,
		&n.Reading,
		&n.Writing,
		&n.Waiting,
	)

	return nil
}

// GetPoints will return exactly one point. The number of bytes read.
func (n *Nginx) GetPoints() []*timeseries.Point {
	points := make([]*timeseries.Point, 7)

	points[0] = plugins.SimplePoint("nginx.ActiveConnections", n.ActiveConnections)
	points[1] = plugins.SimplePoint("nginx.Accepts", n.Accepts)
	points[2] = plugins.SimplePoint("nginx.Handled", n.Handled)
	points[3] = plugins.SimplePoint("nginx.Requests", n.Requests)
	points[4] = plugins.SimplePoint("nginx.Reading", n.Reading)
	points[5] = plugins.SimplePoint("nginx.Writing", n.Writing)
	points[6] = plugins.SimplePoint("nginx.Waiting", n.Waiting)

	return points
}

// GetDoc explains the returned points from GetPoints().
func (n *Nginx) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("Nginx stub status")

	doc.AddMeasurement("nginx.ActiveConnections", "The current number of active client connections including Waiting connections.", "n")
	doc.AddMeasurement("nginx.Accepts", "The total number of accepted client connections.", "n")
	doc.AddMeasurement("nginx.Handled", "The total number of handled connections. Generally, the parameter value is the same as accepts unless some resource limits have been reached (for example, the worker_connections limit).", "n")
	doc.AddMeasurement("nginx.Requests", "The total number of client requests.", "n")
	doc.AddMeasurement("nginx.Reading", "The current number of connections where nginx is reading the request header.", "n")
	doc.AddMeasurement("nginx.Writing", "The current number of connections where nginx is writing the response back to the client.", "n")
	doc.AddMeasurement("nginx.Waiting", "The current number of idle client connections waiting for a request.", "n")

	return doc
}

// Ensure compliance.
var _ plugins.Agent = (*Nginx)(nil)
