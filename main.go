package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/influxdb/influxdb/client"

	"github.com/abrander/agento/configuration"
	"github.com/abrander/agento/plugins"
	_ "github.com/abrander/agento/plugins/cpuspeed"
	_ "github.com/abrander/agento/plugins/cpustats"
	_ "github.com/abrander/agento/plugins/diskstats"
	_ "github.com/abrander/agento/plugins/diskusage"
	_ "github.com/abrander/agento/plugins/entropy"
	"github.com/abrander/agento/plugins/hostname"
	_ "github.com/abrander/agento/plugins/loadstats"
	_ "github.com/abrander/agento/plugins/memorystats"
	_ "github.com/abrander/agento/plugins/netstat"
	_ "github.com/abrander/agento/plugins/openfiles"
	_ "github.com/abrander/agento/plugins/snmpstats"
	_ "github.com/abrander/agento/plugins/socketstats"
)

var config = configuration.Configuration{}

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
		LogError("InfluxDB error: %s", err.Error())
		log.Fatal(err)
	}

	return con
}

func sendToInflux(stats plugins.Results) {
	con := getInfluxClient()
	points := stats.GetPoints()

	// Add hostname tag to all points
	hostname := string(*stats["h"].(*hostname.Hostname))
	for i := range points {
		if points[i].Tags != nil {
			points[i].Tags["hostname"] = hostname
		} else {
			points[i].Tags = map[string]string{"hostname": hostname}
		}
	}

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
			LogWarning("Error writing to influxdb: "+err.Error()+", retry %d/%d", i, 5)
			time.Sleep(time.Millisecond * 500)
			_, err = con.Write(bps)
			if err == nil {
				break
			}
		}
		if i >= retries {
			LogError("Error writing to influxdb: " + err.Error() + ", giving up")
		}
	}
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

func main() {
	if len(os.Args) > 1 && os.Args[1] == "gendoc" {
		d := plugins.GetDoc()

		for _, doc := range d {
			fmt.Printf("| %-40s | %-110s |\n", doc.Description, "")
			fmt.Printf("|------------------------------------------|----------------------------------------------------------------------------------------------------------------|\n")
			for name, description := range doc.Tags {
				fmt.Printf("| %-40s | %-110s |\n", "Tag: "+name, description)
			}
			for name, description := range doc.Measurements {
				fmt.Printf("| %-40s | %-110s |\n", name, description)
			}
			fmt.Printf("\n")
		}
		os.Exit(0)
	}

	err := config.LoadFromFile("/etc/agento.conf")
	InitLogging(&config)

	if err != nil {
		LogError("Configuration error: %s", err.Error())
		os.Exit(1)
	}

	http.HandleFunc("/report", reportHandler)
	http.HandleFunc("/health", healthHandler)

	wg := &sync.WaitGroup{}

	if config.Server.Http.Enabled {
		wg.Add(1)

		go func() {
			// Listen for http connections if needed
			addr := config.Server.Http.Bind + ":" + strconv.Itoa(int(config.Server.Http.Port))
			LogInfo("Listening for http at " + addr)
			err := http.ListenAndServe(addr, nil)
			if err != nil {
				LogError("ListenAndServe(%s): %s", addr, err.Error())
				log.Fatal("ListenAndServe: ", err)
			}

			wg.Done()
		}()
	}

	if config.Server.Https.Enabled {
		wg.Add(1)

		go func() {
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
			LogInfo("Listening for https at " + addr)
			server := &http.Server{Addr: addr, Handler: nil, TLSConfig: tlsConfig}
			err = server.ListenAndServeTLS(config.Server.Https.CertPath, config.Server.Https.KeyPath)
			if err != nil {
				LogError("ListenAndServeTLS(%s): %s", addr, err.Error())
				log.Fatal("ListenAndServe: ", err)
			}

			wg.Done()
		}()
	}

	if config.Server.Udp.Enabled {

		samples := make(chan *Sample)

		// UDP reader loop
		wg.Add(1)
		go func() {
			addr := config.Server.Udp.Bind + ":" + strconv.Itoa(int(config.Server.Udp.Port))

			laddr, err := net.ResolveUDPAddr("udp", addr)
			if err != nil {
				LogError("ResolveUDPAddr(%s): %s", addr, err.Error())
				log.Fatal("ResolveUDPAddr: ", err)
			}

			conn, err := net.ListenUDP("udp", laddr)
			if err != nil {
				LogError("ListenUDP(%s): %s", addr, err.Error())
				log.Fatal("ListenUDP: ", err)
			}

			defer conn.Close()

			buf := make([]byte, 65535)
			var sample Sample

			for {
				n, _, err := conn.ReadFromUDP(buf)

				if err == nil && json.Unmarshal(buf[:n], &sample) == nil {
					samples <- &sample
				}
			}
		}()

		c := time.Tick(time.Second * time.Duration(config.Server.Udp.Interval))

		// Main loop
		wg.Add(1)
		go func() {
			for {
				select {
				case sample := <-samples:
					AddUdpSample(sample)
				case <-c:
					ReportToInfluxdb()
				}
			}
		}()
	}

	if config.Client.Enabled {
		wg.Add(1)

		LogInfo("agento client started, reporting to %s", config.Client.ServerUrl)

		// Randomize our start time to avoid a big cluster reporting at the exact same time
		time.Sleep(time.Second * time.Duration(rand.Intn(config.Client.Interval)))

		// We need to gather one unreported set of metrics. It's needed for
		// calculating deltas on first real report
		plugins.GatherAll()

		c := time.Tick(time.Second * time.Duration(config.Client.Interval))
		for _ = range c {
			results := plugins.GatherAll()
			json, err := json.Marshal(results)

			if err == nil {
				client := &http.Client{}
				req, err := http.NewRequest("POST", config.Client.ServerUrl, bytes.NewReader(json))
				if err != nil {
					LogError(err.Error())
					continue
				}

				if config.Client.Secret != "" {
					req.Header.Add("X-Agento-Secret", config.Client.Secret)
				}

				res, err := client.Do(req)
				if err != nil {
					LogError(err.Error())
					continue
				}
				io.Copy(ioutil.Discard, res.Body)
				res.Body.Close()
			} else {
				LogError(err.Error())
			}

		}
	}

	wg.Wait()
}
