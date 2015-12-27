package openfiles

import (
	"errors"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdb/influxdb/client"

	"github.com/abrander/agento/configuration"
	"github.com/abrander/agento/plugins"
)

func init() {
	plugins.Register("o", NewOpenFiles)
}

func NewOpenFiles() plugins.Plugin {
	return new(OpenFiles)
}

// https://www.kernel.org/doc/Documentation/sysctl/fs.txt

type OpenFiles struct {
	Open int64 `json:"o"`
	Max  int64 `json:"m"`
}

func (stat *OpenFiles) Gather(transport plugins.Transport) error {
	path := filepath.Join(configuration.ProcPath, "/sys/fs/file-nr")
	contents, err := transport.ReadFile(path)
	if err != nil {
		return err
	}

	fields := strings.Fields(string(contents))
	if len(fields) != 3 {
		return errors.New("Unknown format read from " + path)
	}

	stat.Open, _ = strconv.ParseInt(fields[0], 10, 64)
	stat.Max, _ = strconv.ParseInt(fields[2], 10, 64)

	return nil
}

func (o *OpenFiles) GetPoints() []client.Point {
	points := make([]client.Point, 2)

	points[0] = plugins.SimplePoint("misc.OpenFilesUsed", o.Open)
	points[1] = plugins.SimplePoint("misc.OpenFilesFree", o.Max-o.Open)

	return points
}

func (o *OpenFiles) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("File handle usage")

	doc.AddMeasurement("misc.OpenFilesUsed", "The number of allocated file handles", "n")
	doc.AddMeasurement("misc.OpenFilesFree", "The number of free file handles", "n")

	return doc
}

// Ensure compliance
var _ plugins.Agent = (*OpenFiles)(nil)
