package exec

import (
	"bufio"
	"regexp"
	"strconv"

	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/timeseries"
)

func init() {
	plugins.Register("exec", newExec)
}

// Exec will retrieve stub status.
type Exec struct {
	KeyValue []KeyValue `json:"c"`
}

type KeyValue struct {
	key   string
	value float64
}

func newExec() interface{} {
	return new(Exec)
}

// Gather will measure how many bytes can be read from /dev/null.
func (e *Exec) Gather(transport plugins.Transport) error {
	stdout, _, _ := transport.Exec("./plugins/agents/exec/example.sh", "")

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		re := regexp.MustCompile("^(.*).value ([0-9]+(\\.([0-9])*)?)$")
		matches := re.FindAllStringSubmatch(scanner.Text(), -1)

		value, _ := strconv.ParseFloat(matches[0][2], 64)

		kv := KeyValue{}
		kv.key = matches[0][1]
		kv.value = value

		e.KeyValue = append(e.KeyValue, kv)
	}

	return nil
}

// GetPoints will return exactly one point. The number of bytes read.
func (e *Exec) GetPoints() []*timeseries.Point {
	points := make([]*timeseries.Point, len(e.KeyValue))

	for i, kv := range e.KeyValue {
		points[i] = plugins.SimplePoint(kv.key, kv.value)
	}
	return points
}

// GetDoc explains the returned points from GetPoints().
func (e *Exec) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("Exec stub status")

	doc.AddMeasurement("nginx.ActiveConnections", "The current number of active client connections including Waiting connections.", "n")

	return doc
}

// Ensure compliance.
var _ plugins.Agent = (*Exec)(nil)
