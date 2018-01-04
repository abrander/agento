package mysqlslave

import (
	"database/sql"
	"errors"

	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/plugins/agents/mysql"
	"github.com/abrander/agento/timeseries"
)

type Connection struct {
	ConnectionName      string `json:"n"`
	SecondsBehindMaster int64  `json:"sbm"`
	ExecutedLogEntries  int64  `json:"ele"`
}

type MysqlSlave struct {
	Connections []Connection `json:"c"`

	DSN string `toml:"dsn" json:"dsn" description:"Mysql DSN"`
}

func init() {
	plugins.Register("mysqlslave", NewMysqlSlave)
}

func NewMysqlSlave() interface{} {
	return new(MysqlSlave)
}

func (m *MysqlSlave) Gather(transport plugins.Transport) error {
	db, err := mysql.Dial(transport, m.DSN)
	if err != nil {
		return err
	}

	defer db.Close()

	rows, err := db.Query("SHOW ALL SLAVES STATUS")
	if err != nil {
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	row := make([]interface{}, len(columns), len(columns))

	// We use this trickery to find relevant columns. It's ugly.
	nameIndex := -1
	behindIndex := -1
	executedIndex := -1

	for i, name := range columns {
		switch name {
		case "Connection_name":
			nameIndex = i
			row[i] = new(string)

		case "Seconds_Behind_Master":
			behindIndex = i
			row[i] = new(int64)

		case "Executed_log_entries":
			executedIndex = i
			row[i] = new(int64)

		default:
			// This is just to give Scan() something to scan to. One cannot choose nil.
			row[i] = &sql.RawBytes{}
		}
	}

	if behindIndex == -1 || executedIndex == -1 || nameIndex == -1 {
		return errors.New("Something went wrong")
	}

	for rows.Next() {
		conn := Connection{}

		err = rows.Scan(row...)
		if err != nil {
			return err
		}

		conn.ConnectionName = *row[nameIndex].(*string)
		conn.ExecutedLogEntries = *row[executedIndex].(*int64)
		conn.SecondsBehindMaster = *row[behindIndex].(*int64)

		m.Connections = append(m.Connections, conn)
	}

	return nil
}

func (m *MysqlSlave) GetPoints() []*timeseries.Point {
	points := make([]*timeseries.Point, len(m.Connections)*2)

	for i, conn := range m.Connections {
		points[i*2] = plugins.PointWithTag("mysqlslave.SecondsBehindMaster", conn.SecondsBehindMaster, "connection", conn.ConnectionName)
		points[i*2+1] = plugins.PointWithTag("mysqlslave.ExecutedLogEntries", conn.ExecutedLogEntries, "connection", conn.ConnectionName)
	}

	return points
}

func (m *MysqlSlave) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("Mysql slave statistics")

	doc.AddTag("connection", "The connection name from MySQL")
	doc.AddMeasurement("mysqlslave.SecondsBehindMaster", "Difference between the timestamp logged on the master for the event that the slave is currently processing, and the current timestamp on the slave. Zero if the slave is not currently processing an event.", "s")
	doc.AddMeasurement("mysqlslave.ExecutedLogEntries", "How many log entries the slave has executed.", "n")
	return doc
}

// Ensure compliance
var _ plugins.Agent = (*MysqlSlave)(nil)
