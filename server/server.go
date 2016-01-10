package server

import (
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/influxdb/influxdb/client"

	"github.com/abrander/agento/configuration"
	"github.com/abrander/agento/logger"
	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/plugins/agents/hostname"
)

func init() {
	http.HandleFunc("/report", reportHandler)
	http.HandleFunc("/health", healthHandler)
}

var config = configuration.Configuration{}

func Init(cfg configuration.Configuration) {
	config = cfg
}

func getInfluxClient() *client.Client {
	u, _ := url.Parse(config.Server.Influxdb.Url)

	conf := client.Config{
		URL:       *u,
		Username:  config.Server.Influxdb.Username,
		Password:  config.Server.Influxdb.Password,
		UserAgent: "agento-server",
	}

	con, err := client.NewClient(conf)
	if err != nil {
		logger.Red("server", "InfluxDB error: %s", err.Error())
	}

	return con
}

func WritePoints(points []client.Point) error {
	con := getInfluxClient()

	bps := client.BatchPoints{
		Time:            time.Now(),
		Points:          points,
		Database:        config.Server.Influxdb.Database,
		RetentionPolicy: config.Server.Influxdb.RetentionPolicy,
	}
	retries := config.Server.Influxdb.Retries

	_, err := con.Write(bps)
	if err != nil {
		var i int
		for i = 1; i <= retries; i++ {
			logger.Red("server", "Error writing to influxdb: "+err.Error()+", retry %d/%d", i, 5)
			time.Sleep(time.Millisecond * 500)
			_, err = con.Write(bps)
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

func sendToInflux(stats plugins.Results) {
	points := stats.GetPoints()

	// Add hostname tag to all points
	hostname := string(*stats["hostname"].(*hostname.Hostname))
	for i := range points {
		if points[i].Tags != nil {
			points[i].Tags["hostname"] = hostname
		} else {
			points[i].Tags = map[string]string{"hostname": hostname}
		}
	}

	WritePoints(points)
}

func reportHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "Only POST allowed", 400)
		return
	}

	var results = plugins.Results{}

	d := json.NewDecoder(req.Body)
	d.UseNumber()
	err := d.Decode(&results)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	sendToInflux(results)
}

func healthHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Only GET allowed", 400)
		return
	}

	con := getInfluxClient()

	_, _, err := con.Ping()
	if err != nil {
		http.Error(w, "Can't ping InfluxDB", http.StatusServiceUnavailable)
		return
	}

	w.Write([]byte("ok"))
}

func ListenAndServe() {
	// Listen for http connections if needed
	addr := config.Server.Http.Bind + ":" + strconv.Itoa(int(config.Server.Http.Port))
	logger.Yellow("server", "Listening for http at %s", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		logger.Red("ListenAndServe(%s): %s", addr, err.Error())
	}
}

func ListenAndServeTLS() {
	// Listen for https connections if needed
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
	addr := config.Server.Https.Bind + ":" + strconv.Itoa(int(config.Server.Https.Port))
	logger.Yellow("server", "Listening for https at %s", addr)
	server := &http.Server{Addr: addr, Handler: nil, TLSConfig: tlsConfig}
	err := server.ListenAndServeTLS(config.Server.Https.CertPath, config.Server.Https.KeyPath)
	if err != nil {
		logger.Red("ListenAndServeTLS(%s): %s", addr, err.Error())
	}
}

func ListenAndServeUDP() {
	samples := make(chan *Sample)

	// UDP reader loop
	go func() {
		addr := config.Server.Udp.Bind + ":" + strconv.Itoa(int(config.Server.Udp.Port))

		laddr, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			logger.Red("server", "ResolveUDPAddr(%s): %s", addr, err.Error())
			return
		}

		conn, err := net.ListenUDP("udp", laddr)
		if err != nil {
			logger.Red("server", "ListenUDP(%s): %s", addr, err.Error())
			return
		}

		defer conn.Close()

		buf := make([]byte, 65535)

		for {
			var sample Sample
			n, _, err := conn.ReadFromUDP(buf)

			if err == nil && json.Unmarshal(buf[:n], &sample) == nil {
				samples <- &sample
			}
		}
	}()

	c := time.Tick(time.Second * time.Duration(config.Server.Udp.Interval))

	// Main loop
	for {
		select {
		case sample := <-samples:
			AddUdpSample(sample)
		case <-c:
			ReportToInfluxdb()
		}
	}
}
