package timeseries

import (
	"time"

	"github.com/influxdata/influxdb/client/v2"

	"github.com/abrander/agento/configuration"
	"github.com/abrander/agento/logger"
)

type (
	InfluxDb struct {
		conn    client.Client
		retries int
		bpsConf client.BatchPointsConfig
	}
)

func NewInfluxDb(cfg *configuration.InfluxdbConfiguration) (*InfluxDb, error) {
	conf := client.HTTPConfig{
		Addr:      cfg.Url,
		Username:  cfg.Username,
		Password:  cfg.Password,
		UserAgent: "agento-server",
	}

	conn, err := client.NewHTTPClient(conf)
	if err != nil {
		return nil, err
	}

	return &InfluxDb{
		conn:    conn,
		retries: cfg.Retries,
		bpsConf: client.BatchPointsConfig{
			Database:         cfg.Database,
			RetentionPolicy:  cfg.RetentionPolicy,
			WriteConsistency: "one",
		},
	}, nil
}

func (i *InfluxDb) WritePoints(points []*client.Point) error {
	bps, err := client.NewBatchPoints(i.bpsConf)
	if err != nil {
		return err
	}

	for _, point := range points {
		bps.AddPoint(point)
	}

	retries := i.retries

	err = i.conn.Write(bps)
	if err != nil {
		var retry int
		for retry = 1; retry <= retries; retry++ {
			logger.Red("influxdb", "Error writing to influxdb: "+err.Error()+", retry %d/%d", retry, 5)
			time.Sleep(time.Millisecond * 500)
			err = i.conn.Write(bps)
			if err == nil {
				break
			}
		}
		if retry >= retries {
			logger.Red("influxdb", "Error writing to influxdb: "+err.Error()+", giving up")
		}
	}

	return err
}
