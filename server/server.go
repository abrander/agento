package server

import (
	"crypto/tls"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/influxdata/influxdb/client/v2"

	"github.com/abrander/agento/configuration"
	"github.com/abrander/agento/logger"
	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/plugins/agents/hostname"
)

type (
	Server struct {
		influxDb        client.Client
		influxDbRetries int
		inventory       map[string]*Inventory
		bpsConf         client.BatchPointsConfig
		http            configuration.HttpConfiguration
		https           configuration.HttpsConfiguration
		udp             configuration.UdpConfiguration
	}
)

func NewServer(router gin.IRouter, cfg configuration.ServerConfiguration) (*Server, error) {
	s := &Server{}

	router.Any("/report", s.reportHandler)
	router.Any("/health", s.healthHandler)

	conf := client.HTTPConfig{
		Addr:      cfg.Influxdb.Url,
		Username:  cfg.Influxdb.Username,
		Password:  cfg.Influxdb.Password,
		UserAgent: "agento-server",
	}

	var err error
	s.influxDb, err = client.NewHTTPClient(conf)
	if err != nil {
		return nil, err
	}

	s.bpsConf = client.BatchPointsConfig{
		Database:         cfg.Influxdb.Database,
		RetentionPolicy:  cfg.Influxdb.RetentionPolicy,
		WriteConsistency: "1",
	}

	s.influxDbRetries = cfg.Influxdb.Retries

	s.http = cfg.Http
	s.https = cfg.Https
	s.udp = cfg.Udp

	s.inventory = make(map[string]*Inventory)

	return s, nil
}

func (s *Server) WritePoints(points []*client.Point) error {
	bps, err := client.NewBatchPoints(s.bpsConf)
	if err != nil {
		return err
	}

	for _, point := range points {
		bps.AddPoint(point)
	}

	retries := s.influxDbRetries

	err = s.influxDb.Write(bps)
	if err != nil {
		var i int
		for i = 1; i <= retries; i++ {
			logger.Red("server", "Error writing to influxdb: "+err.Error()+", retry %d/%d", i, 5)
			time.Sleep(time.Millisecond * 500)
			err = s.influxDb.Write(bps)
			if err == nil {
				break
			}
		}
		if i >= retries {
			logger.Red("server", "Error writing to influxdb: "+err.Error()+", giving up")
		}
	}

	return err
}

func (s *Server) sendToInflux(stats plugins.Results) error {
	points := stats.GetPoints()

	// Add hostname tag to all points
	hostname := string(*stats["hostname"].(*hostname.Hostname))
	for index, point := range points {
		// FIXME: We do this hack while we wait for InfluxDB PR 5387:
		// https://github.com/influxdata/influxdb/pull/5387
		tags := point.Tags()

		tags["hostname"] = hostname

		points[index], _ = client.NewPoint(point.Name(), tags, point.Fields())
	}

	return s.WritePoints(points)
}

func (s *Server) reportHandler(c *gin.Context) {
	if c.Request.Method != "POST" {
		c.Header("Allow", "POST")
		c.String(http.StatusMethodNotAllowed, "only POST allowed")
		return
	}

	var results = plugins.Results{}

	err := c.BindJSON(&results)
	if err != nil {
		c.String(http.StatusBadRequest, "%s", err.Error())
		return
	}

	err = s.sendToInflux(results)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.String(http.StatusOK, "%s", "Got it")
}

func (s *Server) healthHandler(c *gin.Context) {
	if c.Request.Method != "GET" {
		c.Header("Allow", "GET")
		c.String(http.StatusMethodNotAllowed, "only GET allowed")
		return
	}

	c.String(200, "ok")
}

func (s *Server) ListenAndServe(engine *gin.Engine) {
	addr := s.http.Bind + ":" + strconv.Itoa(int(s.http.Port))

	err := http.ListenAndServe(addr, engine)
	if err != nil {
		logger.Red("server", "ListenAndServe(%s): %s", addr, err.Error())
	} else {
		logger.Yellow("server", "Listening for http at %s", addr)
	}
}

func (s *Server) ListenAndServeTLS(engine *gin.Engine) {
	// Choose strong TLS defaults
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
	}

	addr := s.https.Bind + ":" + strconv.Itoa(int(s.https.Port))

	server := &http.Server{
		Addr:      addr,
		Handler:   engine,
		TLSConfig: tlsConfig}

	err := server.ListenAndServeTLS(s.https.CertPath, s.https.KeyPath)
	if err != nil {
		logger.Red("server", "ListenAndServeTLS(%s): %s", addr, err.Error())
	} else {
		logger.Yellow("server", "Listening for https at %s", addr)
	}
}
