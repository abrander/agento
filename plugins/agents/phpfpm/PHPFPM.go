package phpfpm

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/timeseries"
)

func init() {
	plugins.Register("phpfpm", newPHPFPM)
}

func newPHPFPM() interface{} {
	return &PHPFPM{
		ListenPath: "127.0.0.1:9000",
		StatusPath: "/status",
	}
}

type (
	// PHPFPM will collect metrics from a PHP-FPM pool.
	PHPFPM struct {
		ListenPath          string `toml:"listen" json:"listen" description:"The listen path as configured in PHP-FPM"`
		StatusPath          string `toml:"status" json:"status" description:"The status URI as configured in PHP-FPM"`
		Pool                string `json:"p"`
		AcceptedConnections int64  `json:"ac"`
		ListenQueue         int64  `json:"lq"`
		ListenQueueMax      int64  `json:"lqm"`
		ListenQueueLength   int64  `json:"lql"`
		IdleProcesses       int64  `json:"ip"`
		ActiveProcesses     int64  `json:"ap"`
		MaxActiveProcesses  int64  `json:"map"`
		MaxChildrenReached  int64  `json:"mcr"`
		SlowRequests        int64  `json:"sr"`
	}

	record struct {
		Version       byte
		Type          byte
		RequestID     uint16
		ContentLength uint16
		PaddingLength byte
		Reserved      byte
	}

	appRecord struct {
		Role     uint16
		Flags    byte
		Reserved [5]byte
	}
)

// Gather will connect to PHP-FPM through a tcp or unix socket.
func (a *PHPFPM) Gather(transport plugins.Transport) error {
	var network string

	if strings.HasPrefix(a.ListenPath, "/") {
		network = "unix"
	} else {
		network = "tcp"
	}

	conn, err := net.Dial(network, a.ListenPath)
	if err != nil {
		return err
	}

	// We implement just enough of FastCGI to "GET" the status page. Nothing
	// more. It will probably break in exciting ways.

	// {FCGI_BEGIN_REQUEST, 1, {FCGI_RESPONDER, 0}}
	app := appRecord{
		Role: 1, // FCGI_RESPONDER
	}

	r := record{
		Version:       1,
		Type:          1, // FCGI_BEGIN_REQUEST
		ContentLength: uint16(binary.Size(app)),
	}

	err = binary.Write(conn, binary.BigEndian, r)
	if err != nil {
		return err
	}

	err = binary.Write(conn, binary.BigEndian, app)
	if err != nil {
		return err
	}

	// {FCGI_PARAMS, 1, "\013\002SERVER_PORT80" "\013\016SERVER_ADDR199.170.183.42 ... "}
	var p params = make(map[string]string)

	p["SCRIPT_NAME"] = a.StatusPath
	p["SCRIPT_FILENAME"] = a.StatusPath
	p["REQUEST_METHOD"] = "GET"

	// It is tempting to add 'p["QUERY_STRING"] = "json"' to make parsing easier,
	// but it doesn't really make it any easier. As of now we're pretty cruel
	// when reading the reply. We parse internal FastCGI signalling as if it
	// were ASCII. The json decoder is not as forgiving as ascii parsing.

	size := p.size()
	r = record{
		Version:       1,
		Type:          4, // FCGI_PARAMS
		ContentLength: size,
	}

	err = binary.Write(conn, binary.BigEndian, r)
	if err != nil {
		return err
	}

	err = p.write(conn)
	if err != nil {
		return err
	}

	// {FCGI_PARAMS, 1, ""}
	r = record{
		Version: 1,
		Type:    4, // FCGI_PARAMS
	}
	err = binary.Write(conn, binary.BigEndian, r)
	if err != nil {
		return err
	}

	// {FCGI_STDIN, 1, ""}
	r = record{
		Version: 1,
		Type:    5, // FCGI_STDIN
	}
	err = binary.Write(conn, binary.BigEndian, r)
	if err != nil {
		return err
	}

	// Read reply
	var reply record
	err = binary.Read(conn, binary.BigEndian, &reply)
	if err != nil {
		return err
	}

	// if we see FCGI_STDERR, we assume something is wrong.
	if reply.Type == 7 {
		l := reply.ContentLength
		if l > 1024 {
			l = 1024
		}
		buffer := make([]byte, l)
		n, _ := conn.Read(buffer)

		return fmt.Errorf("Could not get '%s': %s", a.StatusPath, string(buffer[0:n]))
	}

	if reply.ContentLength < 10240 {
		scanner := bufio.NewScanner(conn)
		// Read past headers
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				break
			}
		}

		// Read body
		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.FieldsFunc(line, func(c rune) bool {
				return c == ':'
			})

			if len(fields) == 2 {
				key := strings.TrimSpace(fields[0])
				value := strings.TrimSpace(fields[1])
				valueInt, _ := strconv.ParseInt(value, 10, 64)
				if err != nil {
					continue
				}

				switch key {
				case "pool":
					a.Pool = value
				case "accepted conn":
					a.AcceptedConnections = valueInt
				case "listen queue":
					a.ListenQueue = valueInt
				case "max listen queue":
					a.ListenQueueMax = valueInt
				case "listen queue len":
					a.ListenQueueLength = valueInt
				case "idle processes":
					a.IdleProcesses = valueInt
				case "active processes":
					a.ActiveProcesses = valueInt
				case "max active processes":
					a.MaxActiveProcesses = valueInt
				case "max children reached":
					a.MaxChildrenReached = valueInt
				case "slow requests":
					a.SlowRequests = valueInt
				}
			}
		}
	}
	err = conn.Close()
	if err != nil {
		return err
	}

	return nil
}

// GetPoints will return points suitable for further processing.
func (a *PHPFPM) GetPoints() []*timeseries.Point {
	// If no pool is set, we assume we didn't succeed in gathering metrics and
	// return no points.
	if a.Pool == "" {
		return nil
	}

	points := make([]*timeseries.Point, 9)

	// Each PHP-FPM host can easily run multiple PHP-FPM pools, so we tag all
	// measurements with the pool name.
	points[0] = plugins.PointWithTag("phpfpm.AcceptedConnections", a.AcceptedConnections, "pool", a.Pool)
	points[1] = plugins.PointWithTag("phpfpm.ListenQueue", a.ListenQueue, "pool", a.Pool)
	points[2] = plugins.PointWithTag("phpfpm.ListenQueueMax", a.ListenQueueMax, "pool", a.Pool)
	points[3] = plugins.PointWithTag("phpfpm.ListenQueueLength", a.ListenQueueLength, "pool", a.Pool)
	points[4] = plugins.PointWithTag("phpfpm.IdleProcesses", a.IdleProcesses, "pool", a.Pool)
	points[5] = plugins.PointWithTag("phpfpm.ActiveProcesses", a.ActiveProcesses, "pool", a.Pool)
	points[6] = plugins.PointWithTag("phpfpm.MaxActiveProcesses", a.MaxActiveProcesses, "pool", a.Pool)
	points[7] = plugins.PointWithTag("phpfpm.MaxChildrenReached", a.MaxChildrenReached, "pool", a.Pool)
	points[8] = plugins.PointWithTag("phpfpm.SlowRequests", a.SlowRequests, "pool", a.Pool)

	return points
}

// GetDoc satisfies the plugins.Plugin interface.
func (a *PHPFPM) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("Statistics from PHP-FPM")

	// FIXME: Someone should consult the PHP-FPM source code to actually understand phpfpm.SlowRequests.

	doc.AddTag("pool", "The PHP-FPM pool the metrics was gathered from")

	doc.AddMeasurement("phpfpm.AcceptedConnections", "The number of request accepted by the pool", "n")
	doc.AddMeasurement("phpfpm.ListenQueue", "The number of request in the queue of pending connections. If this number is non-zero, then you better increase number of process FPM can spawn", "n")
	doc.AddMeasurement("phpfpm.ListenQueueMax", "The maximum number of requests in the queue of pending connections since FPM has started", "n")
	doc.AddMeasurement("phpfpm.ListenQueueLength", "The size of the socket queue of pending connections", "n")
	doc.AddMeasurement("phpfpm.IdleProcesses", "The number of idle FPM processes", "n")
	doc.AddMeasurement("phpfpm.ActiveProcesses", "The number of active FPM processes", "n")
	doc.AddMeasurement("phpfpm.MaxActiveProcesses", "The maximum number of active processes since FPM has started", "n")
	doc.AddMeasurement("phpfpm.MaxChildrenReached", "Number of times, the process limit has been reached, when pm tries to start more children. If that value is not zero, then you may need to increase max process limit for your PHP-FPM pool (only relevant for dynamic or ondemand modes)", "n")
	doc.AddMeasurement("phpfpm.SlowRequests", "The number of \"slow\" requests since FPM was started", "n")

	return doc
}

// Ensure compliance
var _ plugins.Agent = (*PHPFPM)(nil)
