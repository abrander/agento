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
	Cmd string `toml:"cmd" json:"cmd" description:"Command to run"`
	Arg string `toml:"arg" json:"arg" description:"Arguments to command"`
	kv  []KeyValue
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
	stdout, _, _ := transport.Exec(e.Cmd, e.Arg)

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		re := regexp.MustCompile("^(.*).value ([0-9]+(\\.([0-9])*)?)$")
		matches := re.FindAllStringSubmatch(scanner.Text(), -1)

		value, _ := strconv.ParseFloat(matches[0][2], 64)

		kv := KeyValue{}
		kv.key = matches[0][1]
		kv.value = value

		e.kv = append(e.kv, kv)
	}

	return nil
}

// GetPoints will return exactly one point. The number of bytes read.
func (e *Exec) GetPoints() []*timeseries.Point {
	points := make([]*timeseries.Point, len(e.kv))

	for i, kv := range e.kv {
		points[i] = plugins.SimplePoint(kv.key, kv.value)
	}
	return points
}

func (e *Exec) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("Exec doesn't have any measurements, but will read munin plugin format and use key and values.")

	return doc
}

// Ensure compliance.
var _ plugins.Agent = (*Exec)(nil)
