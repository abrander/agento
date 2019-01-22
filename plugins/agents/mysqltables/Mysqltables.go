package mysqltables

import (
	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/plugins/agents/mysql"
	"github.com/abrander/agento/timeseries"
)

type Table struct {
	TableSchema   string `json:"ts"`
	TableName     string `json:"tn"`
	TableType     string `json:"tt"`
	Engine        string `json:"e"`
	TableRows     uint64 `json:"tr"`
	AvgRowLength  uint64 `json:"arl"`
	DataLength    uint64 `json:"dl"`
	MaxDataLength uint64 `json:"mdl"`
	IndexLength   uint64 `json:"il"`
	DataFree      uint64 `json:"df"`
}

type MysqlTables struct {
	Tables []Table `json:"t"`

	DSN string `toml:"dsn" json:"dsn" description:"Mysql DSN"`
}

func init() {
	plugins.Register("mysqltables", NewMysqlTables)
}

func NewMysqlTables() interface{} {
	return new(MysqlTables)
}

func (m *MysqlTables) Gather(transport plugins.Transport) error {
	db, err := mysql.Dial(transport, m.DSN)
	if err != nil {
		return err
	}

	defer db.Close()

	rows, err := db.Query("SELECT TABLE_SCHEMA,TABLE_NAME,TABLE_TYPE,ENGINE,IFNULL(TABLE_ROWS, 0) as TABLE_ROWS,AVG_ROW_LENGTH,DATA_LENGTH,MAX_DATA_LENGTH,INDEX_LENGTH,DATA_FREE FROM information_schema.TABLES")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var tableSchema, tableName, tableType, engine string
		var tableRows, avgRowLength, dataLength, maxDataLength, indexLength, dataFree uint64
		err = rows.Scan(&tableSchema, &tableName, &tableType, &engine, &tableRows, &avgRowLength, &dataLength, &maxDataLength, &indexLength, &dataFree)
		if err != nil {
			return err
		}

		table := Table{}
		table.TableSchema = tableSchema
		table.TableName = tableName
		table.TableType = tableType
		table.Engine = engine
		table.TableRows = tableRows
		table.AvgRowLength = avgRowLength
		table.DataLength = dataLength
		table.MaxDataLength = maxDataLength
		table.IndexLength = indexLength
		table.DataFree = dataFree

		m.Tables = append(m.Tables, table)
	}

	return nil
}

func (m *MysqlTables) GetPoints() []*timeseries.Point {
	points := make([]*timeseries.Point, len(m.Tables)*6)
	for i, tables := range m.Tables {
		points[i*6+0] = plugins.PointWithTag("mysqltables.TableRows."+tables.Engine, tables.TableRows, "table", tables.TableSchema+"."+tables.TableName)
		points[i*6+1] = plugins.PointWithTag("mysqltables.AvgRowLength."+tables.Engine, tables.AvgRowLength, "table", tables.TableSchema+"."+tables.TableName)
		points[i*6+2] = plugins.PointWithTag("mysqltables.DataLength."+tables.Engine, tables.DataLength, "table", tables.TableSchema+"."+tables.TableName)
		points[i*6+3] = plugins.PointWithTag("mysqltables.MaxDataLength."+tables.Engine, tables.MaxDataLength, "table", tables.TableSchema+"."+tables.TableName)
		points[i*6+4] = plugins.PointWithTag("mysqltables.IndexLength."+tables.Engine, tables.IndexLength, "table", tables.TableSchema+"."+tables.TableName)
		points[i*6+5] = plugins.PointWithTag("mysqltables.DataFree."+tables.Engine, tables.DataFree, "table", tables.TableSchema+"."+tables.TableName)
	}
	return points
}

func (m *MysqlTables) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("Mysql slave statistics")

	doc.AddTag("connection", "The connection name from MySQL")
	doc.AddMeasurement("mysqltables.TableRows", "The number of rows. Some storage engines, such as MyISAM, store the exact count. For other storage engines, such as InnoDB, this value is an approximation, and may vary from the actual value by as much as 40% to 50%. In such cases, use SELECT COUNT(*) to obtain an accurate count.", "n")
	doc.AddMeasurement("mysqltables.AvgRowLength", "The average row length.", "n")
	doc.AddMeasurement("mysqltables.DataLength", "For MyISAM, DATA_LENGTH is the length of the data file, in bytes. For InnoDB, DATA_LENGTH is the approximate amount of memory allocated for the clustered index, in bytes. Specifically, it is the clustered index size, in pages, multiplied by the InnoDB page size.", "n")
	doc.AddMeasurement("mysqltables.MaxDataLength", "For MyISAM, MAX_DATA_LENGTH is maximum length of the data file. This is the total number of bytes of data that can be stored in the table, given the data pointer size used. Unused for InnoDB.", "n")
	doc.AddMeasurement("mysqltables.IndexLength", "For MyISAM, INDEX_LENGTH is the length of the index file, in bytes. For InnoDB, INDEX_LENGTH is the approximate amount of memory allocated for non-clustered indexes, in bytes. Specifically, it is the sum of non-clustered index sizes, in pages, multiplied by the InnoDB page size.", "n")
	doc.AddMeasurement("mysqltables.DataFree", "The number of allocated but unused bytes. InnoDB tables report the free space of the tablespace to which the table belongs. For a table located in the shared tablespace, this is the free space of the shared tablespace. If you are using multiple tablespaces and the table has its own tablespace, the free space is for only that table. Free space means the number of bytes in completely free extents minus a safety margin. Even if free space displays as 0, it may be possible to insert rows as long as new extents need not be allocated.", "n")

	return doc
}

// Ensure compliance
var _ plugins.Agent = (*MysqlTables)(nil)
