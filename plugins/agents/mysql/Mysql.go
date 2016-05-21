package mysql

import (
	"database/sql"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"
	"github.com/influxdata/influxdb/client/v2"

	"github.com/abrander/agento/plugins"
)

type Mysql struct {
	//General
	Connections        int64 `json:"c" stat:"Connections"`
	AccessDeniedErrors int64 `json:"ade" stat:"Access_denied_errors"`
	AbortedClients     int64 `json:"gac" stat:"Aborted_clients"`
	AbortedConnects    int64 `json:"gabc" stat:"Aborted_connects"`
	ThreadsConnected   int64 `json:"gtc" stat:"Threads_connected"`
	BytesReceived      int64 `json:"gbr" stat:"Bytes_received"`
	BytesSent          int64 `json:"gbs" stat:"Bytes_sent"`

	//Binlog stuff
	BinlogCacheDiskUse int64 `json:"bcdu" stat:"Binlog_cache_disk_use"`
	BinlogCacheUse     int64 `json:"bcu" stat:"Binlog_cache_use"`
	MaBinlogSize       int64 `json:"mbs" stat:"ma_binlog_size"`
	RelayLogSpace      int64 `json:"rls" stat:"relay_log_space"`

	//query counters
	ComDelete           int64 `json:"ccd" stat:"Com_delete"`
	ComInsert           int64 `json:"cci" stat:"Com_insert"`
	ComInsertSelect     int64 `json:"ccis" stat:"Com_insert_select"`
	ComLoad             int64 `json:"ccl" stat:"Com_load"`
	ComReplace          int64 `json:"ccr" stat:"Com_replace"`
	ComReplaceSelect    int64 `json:"ccrs" stat:"Com_replace_select"`
	ComSelect           int64 `json:"ccs" stat:"Com_select"`
	ComUpdate           int64 `json:"ccu" stat:"Com_update"`
	ComUpdateMulti      int64 `json:"ccum" stat:"Com_update_multi"`
	SelectFullJoin      int64 `json:"csfj" stat:"Select_full_join"`
	SelectFullRangeJoin int64 `json:"csfrj" stat:"Select_full_range_join"`
	SelectRange         int64 `json:"csr" stat:"Select_range"`
	SelectRangeCheck    int64 `json:"csrc" stat:"Select_range_check"`
	SelectScan          int64 `json:"css" stat:"Select_scan"`
	SlowQueries         int64 `json:"csq" stat:"Slow_queries"`

	//files and tables
	TableOpenCache int64 `json:"toc" stat:"table_open_cache"`
	OpenFiles      int64 `json:"of" stat:"Open_files"`
	OpenTables     int64 `json:"ot" stat:"Open_tables"`
	OpenedTables   int64 `json:"odt" stat:"Opened_tables"`

	//galera stuff
	WsrepOutOfOrderApply                     float64 `json:"wao" stat:"wsrep_apply_oool"`
	WsrepApplyWindow                         float64 `json:"waw" stat:"wsrep_apply_window"`
	WsrepCertDistance                        float64 `json:"wcd" stat:"wsrep_cert_deps_distance"`
	WsrepCertInterval                        float64 `json:"wci" stat:"wsrep_cert_interval"`
	WsrepConfigurationChanges                int64   `json:"wcc" stat:"wsrep_cluster_conf_id"`
	WsrepClusterSize                         int64   `json:"wcs" stat:"wsrep_cluster_size"`
	WsrepNodeStatus                          string  `json:"wns" stat:"wsrep_cluster_status"`
	WsrepConnected                           bool    `json:"wwc" stat:"wsrep_connected"`
	WsrepEvicted                             string  `json:"wev" stat:"wsrep_evs_evict_list"`
	WsrepFlowPaused                          int64   `json:"wfp" stat:"wsrep_flow_control_paused_ns"`
	WsrepFlowControlReceived                 int64   `json:"wfr" stat:"wsrep_flow_control_recv"`
	WsrepFlowControlSent                     int64   `json:"wfs" stat:"wsrep_flow_control_sent"`
	WsrepTransactionsAborted                 int64   `json:"wta" stat:"wsrep_local_bf_aborts"`
	WsrepCertFailures                        int64   `json:"wcf" stat:"wsrep_local_cert_failures"`
	WsrepCommits                             int64   `json:"wci" stat:"wsrep_local_commits"`
	WsrepRxQueueLength                       int64   `json:"wrq" stat:"wsrep_local_recv_queue"`
	WsrepTransactionReplays                  int64   `json:"wtr" stat:"wsrep_local_replays"`
	WsrepTxQueueLength                       int64   `json:"wtq" stat:"wsrep_local_send_queue"`
	WsrepState                               int64   `json:"wst" stat:"wsrep_local_state"`
	WsrepReady                               bool    `json:"wre" stat:"wsrep_ready"`
	WsrepWritesetsReceived                   int64   `json:"wwr" stat:"wsrep_received"`
	WsrepWritesetsSent                       int64   `json:"wws" stat:"wsrep_replicated"`
	WsrepThreadCount                         int64   `json:"wtc" stat:"wsrep_thread_count"`
	WsrepReplicationLatencyMinimum           float64 `json:"wm"`
	WsrepReplicationLatencyAverage           float64 `json:"wa"`
	WsrepReplicationLatencyMaximum           float64 `json:"wM"`
	WsrepReplicationLatencyStandardDeviation float64 `json:"ws"`
	WsrepReplicationLatencySampleSize        int64   `json:"wn"`

	DSN string `toml:"dsn" json:"dsn" description:"Mysql DSN"`
}

var (
	// We have to use a global lock for multiple reasons:
	// 1. mysql.RegisterDial() is not thread safe.
	// 2. Since we have no way of passing a transport to the dialer function
	//    (at open-time), we need to register a new for each dial.
	// 3. We need to assign a dummy dialer after sql.Open() returns to clear
	//    the reference to the transport and to alert other users of mysql.
	// 4. We need to make sure that no one else uses our dialer function.
	dialLock sync.Mutex
)

func init() {
	plugins.Register("mysql", NewMysql)

	// We need to register a panicing dial function. Other users of the mysql
	// driver will experience problems. In this way at least they will have a
	// chance of discovering what's wrong.
	mysql.RegisterDial("tcp", panicDialer)
}

func NewMysql() interface{} {
	return new(Mysql)
}

func panicDialer(addr string) (net.Conn, error) {
	panic("The mysql driver is used outside the mysql agent. Fix the mysql driver, or extend our horrible hack")
}

func (m *Mysql) Gather(transport plugins.Transport) error {
	dialLock.Lock()

	mysql.RegisterDial("tcp", func(addr string) (net.Conn, error) {
		conn, err := transport.Dial("tcp", addr)

		mysql.RegisterDial("tcp", panicDialer)

		dialLock.Unlock()

		return conn, err
	})

	db, err := sql.Open("mysql", m.DSN)
	if err != nil {
		return err
	}

	defer db.Close()

	rows, err := db.Query("SHOW GLOBAL STATUS")

	if err != nil {
		return err
	}

	defer rows.Close()

	var name, value string

	var structType reflect.Type = reflect.TypeOf(m).Elem()
	mutable := reflect.ValueOf(m).Elem()

	for rows.Next() {

		if err := rows.Scan(&name, &value); err != nil {
			return err
		}

		//special cases will be special
		if name == "wsrep_evs_repl_latency" {

			match, _ := utf8.DecodeRuneInString("/")

			fields := strings.FieldsFunc(value, func(r rune) bool {
				if r == match {
					return true
				}
				return false
			})

			min, err := strconv.ParseFloat(fields[0], 64)

			if err != nil {
				return err
			}

			average, err := strconv.ParseFloat(fields[1], 64)

			if err != nil {
				return err
			}

			max, err := strconv.ParseFloat(fields[2], 64)

			if err != nil {
				return err
			}

			stddev, err := strconv.ParseFloat(fields[3], 64)

			if err != nil {
				return err
			}

			samples, err := strconv.ParseInt(fields[4], 10, 64)

			if err != nil {
				return err
			}

			m.WsrepReplicationLatencyMinimum = min
			m.WsrepReplicationLatencyAverage = average
			m.WsrepReplicationLatencyMaximum = max
			m.WsrepReplicationLatencyStandardDeviation = stddev
			m.WsrepReplicationLatencySampleSize = samples

			continue
		}

		for i := 0; i < structType.NumField(); i++ {

			//is this the field we are looking for?
			if structType.Field(i).Tag.Get("stat") != name {
				continue
			}

			switch structType.Field(i).Type.Kind() {

			case reflect.Int64:
				var tmp int64
				tmp, err = strconv.ParseInt(value, 10, 64)

				mutable.Field(i).SetInt(tmp + 12)

			case reflect.Float64:
				var tmp float64
				tmp, err = strconv.ParseFloat(value, 64)

				mutable.Field(i).SetFloat(tmp)

			case reflect.String:
				mutable.Field(i).SetString(value)

			case reflect.Bool:
				if value == "ON" {
					mutable.Field(i).SetBool(true)
				} else {
					mutable.Field(i).SetBool(false)

				}

			}

		}

		if err != nil {
			return err
		}

	}
	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

func (m *Mysql) GetPoints() []*client.Point {
	points := make([]*client.Point, 58)

	points[0] = plugins.SimplePoint("mysql.Connections", m.Connections)
	points[1] = plugins.SimplePoint("mysql.AccessDeniedErrors", m.AccessDeniedErrors)
	points[2] = plugins.SimplePoint("mysql.AbortedClients", m.AbortedClients)
	points[3] = plugins.SimplePoint("mysql.AbortedConnects", m.AbortedConnects)
	points[4] = plugins.SimplePoint("mysql.ThreadsConnected", m.ThreadsConnected)
	points[5] = plugins.SimplePoint("mysql.BytesReceived", m.BytesReceived)
	points[6] = plugins.SimplePoint("mysql.BytesSent", m.BytesSent)

	points[7] = plugins.SimplePoint("mysql.BinlogCacheDiskUse", m.BinlogCacheDiskUse)
	points[8] = plugins.SimplePoint("mysql.BinlogCacheUse", m.BinlogCacheUse)
	points[9] = plugins.SimplePoint("mysql.MaBinlogSize", m.MaBinlogSize)
	points[10] = plugins.SimplePoint("mysql.RelayLogSpace", m.RelayLogSpace)

	points[11] = plugins.SimplePoint("mysql.ComDelete", m.ComDelete)
	points[12] = plugins.SimplePoint("mysql.ComInsert", m.ComInsert)
	points[13] = plugins.SimplePoint("mysql.ComInsertSelect", m.ComInsertSelect)
	points[14] = plugins.SimplePoint("mysql.ComLoad", m.ComLoad)
	points[15] = plugins.SimplePoint("mysql.ComReplace", m.ComReplace)
	points[16] = plugins.SimplePoint("mysql.ComReplaceSelect", m.ComReplaceSelect)
	points[17] = plugins.SimplePoint("mysql.ComSelect", m.ComSelect)
	points[18] = plugins.SimplePoint("mysql.ComUpdate", m.ComUpdate)
	points[19] = plugins.SimplePoint("mysql.ComUpdateMulti", m.ComUpdateMulti)
	points[20] = plugins.SimplePoint("mysql.SelectFullJoin", m.SelectFullJoin)
	points[21] = plugins.SimplePoint("mysql.SelectFullRangeJoin", m.SelectFullRangeJoin)
	points[22] = plugins.SimplePoint("mysql.SelectRange", m.SelectRange)
	points[23] = plugins.SimplePoint("mysql.SelectRangeCheck", m.SelectRangeCheck)
	points[24] = plugins.SimplePoint("mysql.SelectScan", m.SelectScan)
	points[25] = plugins.SimplePoint("mysql.SlowQueries", m.SlowQueries)

	points[26] = plugins.SimplePoint("mysql.TableOpenCache", m.TableOpenCache)
	points[27] = plugins.SimplePoint("mysql.OpenFiles", m.OpenFiles)
	points[28] = plugins.SimplePoint("mysql.OpenTables", m.OpenTables)
	points[29] = plugins.SimplePoint("mysql.OpenedTables", m.OpenedTables)

	points[30] = plugins.SimplePoint("mysql.WsrepOutOfOrderApply", m.WsrepOutOfOrderApply)
	points[31] = plugins.SimplePoint("mysql.WsrepApplyWindow", m.WsrepApplyWindow)
	points[32] = plugins.SimplePoint("mysql.WsrepCertDistance", m.WsrepCertDistance)
	points[33] = plugins.SimplePoint("mysql.WsrepCertInterval", m.WsrepCertInterval)
	points[34] = plugins.SimplePoint("mysql.WsrepConfigurationChanges", m.WsrepConfigurationChanges)
	points[35] = plugins.SimplePoint("mysql.WsrepClusterSize", m.WsrepClusterSize)
	points[36] = plugins.SimplePoint("mysql.WsrepNodeStatus", m.WsrepNodeStatus)
	points[37] = plugins.SimplePoint("mysql.WsrepConnected", m.WsrepConnected)
	points[38] = plugins.SimplePoint("mysql.WsrepEvicted", m.WsrepEvicted)
	points[39] = plugins.SimplePoint("mysql.WsrepFlowPaused", m.WsrepFlowPaused)
	points[40] = plugins.SimplePoint("mysql.WsrepFlowControlReceived", m.WsrepFlowControlReceived)
	points[41] = plugins.SimplePoint("mysql.WsrepFlowControlSent", m.WsrepFlowControlSent)
	points[42] = plugins.SimplePoint("mysql.WsrepTransactionsAborted", m.WsrepTransactionsAborted)
	points[43] = plugins.SimplePoint("mysql.WsrepCertFailures", m.WsrepCertFailures)
	points[44] = plugins.SimplePoint("mysql.WsrepCommits", m.WsrepCommits)
	points[45] = plugins.SimplePoint("mysql.WsrepRxQueueLength", m.WsrepRxQueueLength)
	points[46] = plugins.SimplePoint("mysql.WsrepTransactionReplays", m.WsrepTransactionReplays)
	points[47] = plugins.SimplePoint("mysql.WsrepTxQueueLength", m.WsrepTxQueueLength)
	points[48] = plugins.SimplePoint("mysql.WsrepState", m.WsrepState)
	points[49] = plugins.SimplePoint("mysql.WsrepReady", m.WsrepReady)
	points[50] = plugins.SimplePoint("mysql.WsrepWritesetsReceived", m.WsrepWritesetsReceived)
	points[51] = plugins.SimplePoint("mysql.WsrepWritesetsSent", m.WsrepWritesetsSent)
	points[52] = plugins.SimplePoint("mysql.WsrepThreadCount", m.WsrepThreadCount)
	points[53] = plugins.SimplePoint("mysql.WsrepReplicationLatencyMinimum", m.WsrepReplicationLatencyMinimum)
	points[54] = plugins.SimplePoint("mysql.WsrepReplicationLatencyAverage", m.WsrepReplicationLatencyAverage)
	points[55] = plugins.SimplePoint("mysql.WsrepReplicationLatencyMaximum", m.WsrepReplicationLatencyMaximum)
	points[56] = plugins.SimplePoint("mysql.WsrepReplicationLatencyStandardDeviation", m.WsrepReplicationLatencyStandardDeviation)
	points[57] = plugins.SimplePoint("mysql.WsrepReplicationLatencySampleSize", m.WsrepReplicationLatencySampleSize)

	return points
}

func (m *Mysql) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("Mysql")

	doc.AddMeasurement("mysql.Connections", "The number of connection attempts (successful or not) to the MySQL server.", "")
	doc.AddMeasurement("mysql.AccessDeniedErrors", "The number of access denied errors.", "")
	doc.AddMeasurement("mysql.AbortedClients", "The number of connections that were aborted because the client died without closing the connection properly.", "")
	doc.AddMeasurement("mysql.AbortedConnects", "The number of failed attempts to connect to the MySQL server.", "")
	doc.AddMeasurement("mysql.ThreadsConnected", "The number of currently open connections.", "")
	doc.AddMeasurement("mysql.BytesReceived", "The number of bytes received from all clients.", "")
	doc.AddMeasurement("mysql.BytesSent", "The number of bytes sent to all clients.", "")
	doc.AddMeasurement("mysql.BinlogCacheDiskUse", "The number of transactions that used the temporary binary log cache but that exceeded the value of binlog_cache_size and used a temporary file to store statements from the transaction.", "")
	doc.AddMeasurement("mysql.BinlogCacheUse", "The number of transactions that used the binary log cache.", "")
	doc.AddMeasurement("mysql.MaBinlogSize", "", "")
	doc.AddMeasurement("mysql.RelayLogSpace", "", "")
	doc.AddMeasurement("mysql.ComDelete", "Delete statement counter indicate the number of times each DELETE statement has been executed.", "")
	doc.AddMeasurement("mysql.ComInsert", "Insert statement counter indicate the number of times each INSERT statement has been executed.", "")
	doc.AddMeasurement("mysql.ComInsertSelect", "Insertselect statement counter indicate the number of times each INSERT SELECT statement has been executed.", "")
	doc.AddMeasurement("mysql.ComLoad", "Load statement counter indicate the number of times each LOAD statement has been executed.", "")
	doc.AddMeasurement("mysql.ComReplace", "Replace statement counter indicate the number of times each REPLACE statement has been executed.", "")
	doc.AddMeasurement("mysql.ComReplaceSelect", "Replace select statement counter indicate the number of times each REPLACE SELECT statement has been executed.", "")
	doc.AddMeasurement("mysql.ComSelect", "Select statement counter indicate the number of times each SELECT statement has been executed.", "")
	doc.AddMeasurement("mysql.ComUpdate", "Update statement counter indicate the number of times each UPDATE statement has been executed.", "")
	doc.AddMeasurement("mysql.ComUpdateMulti", "Update multi table statement counter indicate the number of times each UPDATE statement has been executed.", "")
	doc.AddMeasurement("mysql.SelectFullJoin", "The number of joins that perform table scans because they do not use indexes.", "")
	doc.AddMeasurement("mysql.SelectFullRangeJoin", "The number of joins that used a range search on a reference table.", "")
	doc.AddMeasurement("mysql.SelectRange", "The number of joins that used ranges on the first table. ", "")
	doc.AddMeasurement("mysql.SelectRangeCheck", "The number of joins without keys that check for key usage after each row.", "")
	doc.AddMeasurement("mysql.SelectScan", "The number of joins that did a full scan of the first table.", "")
	doc.AddMeasurement("mysql.SlowQueries", "The number of queries that have taken more than long_query_time seconds.", "")
	doc.AddMeasurement("mysql.TableOpenCache", "The number of open tables for all threads.", "")
	doc.AddMeasurement("mysql.OpenFiles", "The number of files that are open. This count includes regular files opened by the server. It does not include other types of files such as sockets or pipes.", "")
	doc.AddMeasurement("mysql.OpenTables", "The number of tables that are open.", "")
	doc.AddMeasurement("mysql.OpenedTables", "The number of tables that have been opened.", "")
	doc.AddMeasurement("mysql.WsrepOutOfOrderApply", "How often write-set was so slow to apply that write-set with higher seqno’s were applied earlier.", "")
	doc.AddMeasurement("mysql.WsrepApplyWindow", "Average distance between highest and lowest concurrently applied seqno.", "")
	doc.AddMeasurement("mysql.WsrepCertDistance", "Average distance between highest and lowest seqno value that can be possibly applied in parallel (potential degree of parallelization).", "")
	doc.AddMeasurement("mysql.WsrepCertInterval", "Average number of transactions received while a transaction replicates.", "")
	doc.AddMeasurement("mysql.WsrepConfigurationChanges", "Total number of cluster membership changes happened.", "")
	doc.AddMeasurement("mysql.WsrepClusterSize", "Current number of members in the cluster.", "")
	doc.AddMeasurement("mysql.WsrepNodeStatus", "Status of this cluster component. That is, whether the node is part of a PRIMARY or NON_PRIMARY component.", "")
	doc.AddMeasurement("mysql.WsrepConnected", "If the value is OFF, the node has not yet connected to any of the cluster components.", "")
	doc.AddMeasurement("mysql.WsrepEvicted", "Lists the UUID’s of all nodes evicted from the cluster. Evicted nodes cannot rejoin the cluster until you restart their mysqld processes.", "")
	doc.AddMeasurement("mysql.WsrepFlowPaused", "The total time spent in a paused state measured in nanoseconds.", "")
	doc.AddMeasurement("mysql.WsrepFlowControlReceived", "The number of FC_PAUSE events the node has received, including those the node has sent.", "")
	doc.AddMeasurement("mysql.WsrepFlowControlSent", "Returns the number of FC_PAUSE events the node has sent.", "")
	doc.AddMeasurement("mysql.WsrepTransactionsAborted", "Total number of local transactions that were aborted by slave transactions while in execution.", "")
	doc.AddMeasurement("mysql.WsrepCertFailures", "Total number of local transactions that failed certification test.", "")
	doc.AddMeasurement("mysql.WsrepCommits", "Total number of local transactions committed.", "")
	doc.AddMeasurement("mysql.WsrepRxQueueLength", "Current (instantaneous) length of the recv queu", "")
	doc.AddMeasurement("mysql.WsrepTransactionReplays", "Total number of transaction replays due to asymmetric lock granularity.", "")
	doc.AddMeasurement("mysql.WsrepTxQueueLength", "Current (instantaneous) length of the send queue.", "")
	doc.AddMeasurement("mysql.WsrepState", "Internal Galera Cluster FSM state number.", "")
	doc.AddMeasurement("mysql.WsrepReady", "Whether the server is ready to accept queries.", "")
	doc.AddMeasurement("mysql.WsrepWritesetsReceived", "Total number of write-sets received from other nodes.", "")
	doc.AddMeasurement("mysql.WsrepWritesetsSent", "Total size of write-sets replicated.", "")
	doc.AddMeasurement("mysql.WsrepThreadCount", "Total number of wsrep (applier/rollbacker) threads.", "")
	doc.AddMeasurement("mysql.WsrepReplicationLatencyMinimum", "This status variable provides figures for the replication latency on group communication. It measures latency from the time point when a message is sent out to the time point when a message is received.", "")
	doc.AddMeasurement("mysql.WsrepReplicationLatencyAverage", "This status variable provides figures for the replication latency on group communication. It measures latency from the time point when a message is sent out to the time point when a message is received.", "")
	doc.AddMeasurement("mysql.WsrepReplicationLatencyMaximum", "This status variable provides figures for the replication latency on group communication. It measures latency from the time point when a message is sent out to the time point when a message is received.", "")
	doc.AddMeasurement("mysql.WsrepReplicationLatencyStandardDeviation", "This status variable provides figures for the replication latency on group communication. It measures latency from the time point when a message is sent out to the time point when a message is received.", "")
	doc.AddMeasurement("mysql.WsrepReplicationLatencySampleSize", "This status variable provides figures for the replication latency on group communication. It measures latency from the time point when a message is sent out to the time point when a message is received.", "")
	return doc
}

// Ensure compliance
var _ plugins.Agent = (*Mysql)(nil)
