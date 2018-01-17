package exec

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"sync"

	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/timeseries"
)

func init() {
	plugins.Register("exec", newExec)
}

// Exec will retrieve stub status.
type Exec struct {
	KeyValue []KeyValue `json:"c"`
	wg       sync.WaitGroup
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
	e.wg.Add(1)
	cmd := exec.Command("./plugins/agents/exec/example.sh", "")
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {

		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		defer e.wg.Done()
		for scanner.Scan() {
			re := regexp.MustCompile("^(.*).value ([0-9]+(\\.([0-9])*)?)$")
			matches := re.FindAllStringSubmatch(scanner.Text(), -1)

			value, err2 := strconv.ParseFloat(matches[0][2], 64)
			if err2 != nil {
				//
			}

			kv := KeyValue{}
			kv.key = matches[0][1]
			kv.value = value

			e.KeyValue = append(e.KeyValue, kv)
		}
	}()

	err = cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
		os.Exit(1)
	}

	err = cmd.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error waiting for Cmd", err)
		os.Exit(1)
	}

	return nil
}

// GetPoints will return exactly one point. The number of bytes read.
func (e *Exec) GetPoints() []*timeseries.Point {
	e.wg.Wait()
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
