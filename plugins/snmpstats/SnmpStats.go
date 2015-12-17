package snmpstats

import (
	"bufio"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/influxdb/influxdb/client"

	"github.com/abrander/agento/configuration"
	"github.com/abrander/agento/plugins"
)

func init() {
	plugins.Register("s", NewSnmpStats)
}

func NewSnmpStats() plugins.Plugin {
	return new(SnmpStats)
}

type SnmpStats struct {
	sampletime        time.Time `json:"-"`
	previousSnmpStats *SnmpStats

	IpForwarding      float64 `json:"-" row:"1" col:"1"`
	IpDefaultTTL      float64 `json:"-" row:"1" col:"2"`
	IpInReceives      float64 `json:"-" row:"1" col:"3"`
	IpInHdrErrors     float64 `json:"-" row:"1" col:"4"`
	IpInAddrErrors    float64 `json:"-" row:"1" col:"5"`
	IpForwDatagrams   float64 `json:"-" row:"1" col:"6"`
	IpInUnknownProtos float64 `json:"-" row:"1" col:"7"`
	IpInDiscards      float64 `json:"-" row:"1" col:"8"`
	IpInDelivers      float64 `json:"-" row:"1" col:"9"`
	IpOutRequests     float64 `json:"-" row:"1" col:"10"`
	IpOutDiscards     float64 `json:"-" row:"1" col:"11"`
	IpOutNoRoutes     float64 `json:"-" row:"1" col:"12"`
	IpReasmTimeout    float64 `json:"-" row:"1" col:"13"`
	IpReasmReqds      float64 `json:"-" row:"1" col:"14"`
	IpReasmOKs        float64 `json:"-" row:"1" col:"15"`
	IpReasmFails      float64 `json:"-" row:"1" col:"16"`
	IpFragOKs         float64 `json:"-" row:"1" col:"17"`
	IpFragFails       float64 `json:"-" row:"1" col:"18"`
	IpFragCreates     float64 `json:"-" row:"1" col:"19"`

	IcmpInMsgs           float64 `json:"-" row:"3" col:"1"`
	IcmpInErrors         float64 `json:"-" row:"3" col:"2"`
	IcmpInCsumErrors     float64 `json:"-" row:"3" col:"3"`
	IcmpInDestUnreachs   float64 `json:"-" row:"3" col:"4"`
	IcmpInTimeExcds      float64 `json:"-" row:"3" col:"5"`
	IcmpInParmProbs      float64 `json:"-" row:"3" col:"6"`
	IcmpInSrcQuenchs     float64 `json:"-" row:"3" col:"7"`
	IcmpInRedirects      float64 `json:"-" row:"3" col:"8"`
	IcmpInEchos          float64 `json:"-" row:"3" col:"9"`
	IcmpInEchoReps       float64 `json:"-" row:"3" col:"10"`
	IcmpInTimestamps     float64 `json:"-" row:"3" col:"11"`
	IcmpInTimestampReps  float64 `json:"-" row:"3" col:"12"`
	IcmpInAddrMasks      float64 `json:"-" row:"3" col:"13"`
	IcmpInAddrMaskReps   float64 `json:"-" row:"3" col:"14"`
	IcmpOutMsgs          float64 `json:"-" row:"3" col:"15"`
	IcmpOutErrors        float64 `json:"-" row:"3" col:"16"`
	IcmpOutDestUnreachs  float64 `json:"-" row:"3" col:"17"`
	IcmpOutTimeExcds     float64 `json:"-" row:"3" col:"18"`
	IcmpOutParmProbs     float64 `json:"-" row:"3" col:"19"`
	IcmpOutSrcQuenchs    float64 `json:"-" row:"3" col:"20"`
	IcmpOutRedirects     float64 `json:"-" row:"3" col:"21"`
	IcmpOutEchos         float64 `json:"-" row:"3" col:"22"`
	IcmpOutEchoReps      float64 `json:"-" row:"3" col:"23"`
	IcmpOutTimestamps    float64 `json:"-" row:"3" col:"24"`
	IcmpOutTimestampReps float64 `json:"-" row:"3" col:"25"`
	IcmpOutAddrMasks     float64 `json:"-" row:"3" col:"26"`
	IcmpOutAddrMaskReps  float64 `json:"-" row:"3" col:"27"`

	IcmpMsgInType0  float64 `json:"-" row:"5" col:"1"`
	IcmpMsgInType3  float64 `json:"-" row:"5" col:"2"`
	IcmpMsgOutType3 float64 `json:"-" row:"5" col:"3"`
	IcmpMsgOutType8 float64 `json:"-" row:"5" col:"4"`

	TcpRtoAlgorithm float64 `json:"-" row:"7" col:"1"`
	TcpRtoMin       float64 `json:"-" row:"7" col:"2"`
	TcpRtoMax       float64 `json:"-" row:"7" col:"3"`
	TcpMaxConn      float64 `json:"-" row:"7" col:"4"`
	TcpActiveOpens  float64 `json:"-" row:"7" col:"5"`
	TcpPassiveOpens float64 `json:"-" row:"7" col:"6"`
	TcpAttemptFails float64 `json:"-" row:"7" col:"7"`
	TcpEstabResets  float64 `json:"-" row:"7" col:"8"`
	TcpCurrEstab    float64 `json:"-" row:"7" col:"9"`
	TcpInSegs       float64 `json:"-" row:"7" col:"10"`
	TcpOutSegs      float64 `json:"-" row:"7" col:"11"`
	TcpRetransSegs  float64 `json:"-" row:"7" col:"12"`
	TcpInErrs       float64 `json:"-" row:"7" col:"13"`
	TcpOutRsts      float64 `json:"-" row:"7" col:"14"`
	TcpInCsumErrors float64 `json:"-" row:"7" col:"15"`

	UdpInDatagrams  float64 `json:"-" row:"9" col:"1"`
	UdpNoPorts      float64 `json:"-" row:"9" col:"2"`
	UdpInErrors     float64 `json:"-" row:"9" col:"3"`
	UdpOutDatagrams float64 `json:"-" row:"9" col:"4"`
	UdpRcvbufErrors float64 `json:"-" row:"9" col:"5"`
	UdpSndbufErrors float64 `json:"-" row:"9" col:"6"`
	UdpInCsumErrors float64 `json:"-" row:"9" col:"7"`

	UdpLiteInDatagrams  float64 `json:"-" row:"11" col:"1"`
	UdpLiteNoPorts      float64 `json:"-" row:"11" col:"2"`
	UdpLiteInErrors     float64 `json:"-" row:"11" col:"3"`
	UdpLiteOutDatagrams float64 `json:"-" row:"11" col:"4"`
	UdpLiteRcvbufErrors float64 `json:"-" row:"11" col:"5"`
	UdpLiteSndbufErrors float64 `json:"-" row:"11" col:"6"`
	UdpLiteInCsumErrors float64 `json:"-" row:"11" col:"7"`
}

func (snmp *SnmpStats) Gather() error {
	stat := SnmpStats{}

	path := filepath.Join(configuration.ProcPath, "/net/snmp")
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	stat.sampletime = time.Now()
	scanner := bufio.NewScanner(file)
	var a [12][]string
	row := 0
	for scanner.Scan() {
		text := scanner.Text()

		a[row] = strings.Fields(strings.Trim(text, " "))

		row += 1
	}

	v := reflect.TypeOf(stat)
	s := reflect.ValueOf(&stat).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		srow := field.Tag.Get("row")
		scol := field.Tag.Get("col")

		if srow != "" && scol != "" {
			row, _ = strconv.Atoi(srow)
			col, _ := strconv.Atoi(scol)

			if row >= 0 && row < len(a) && col >= 0 && col < len(a[row]) {
				value, err := strconv.ParseInt(a[row][col], 10, 64)

				if err == nil {
					s.Field(i).SetFloat(float64(value))
				}
			}
		}
	}

	*snmp = *stat.Sub(snmp.previousSnmpStats)
	snmp.previousSnmpStats = &stat

	return nil
}

func (c *SnmpStats) Sub(previousSnmpStats *SnmpStats) *SnmpStats {
	// FIXME: Compute delta

	return c
}

func (s *SnmpStats) GetPoints() []client.Point {
	points := make([]client.Point, 0)

	// FIXME: Return something ;)

	return points
}

func (c *SnmpStats) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc()

	// FIXME: Return something ;)

	return doc
}
