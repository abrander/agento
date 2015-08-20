package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/influxdb/influxdb/client"

	"agento"
)

var config = agento.Configuration{}

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
		agento.LogError("InfluxDB error: %s", err.Error())
		log.Fatal(err)
	}

	return con
}

func sendToInflux(stats agento.MachineStats) {
	con := getInfluxClient()
	points := stats.GetPoints()

	// Add hostname tag to all points
	for i := range points {
		if points[i].Tags != nil {
			points[i].Tags["hostname"] = stats.Hostname
		} else {
			points[i].Tags = map[string]string{"hostname": stats.Hostname}
		}
	}

	bps := client.BatchPoints{
		Time:            time.Now(),
		Points:          points,
		Database:        config.Server.Influxdb.Database,
		RetentionPolicy: config.Server.Influxdb.RetentionPolicy,
	}

	_, err := con.Write(bps)
	if err != nil {
		for i := 1; i <= 5; i++ {
			agento.LogWarning("Error writing to influxdb: "+err.Error()+", retry %d/%d", i, 5)
			time.Sleep(time.Millisecond * 500)
			_, err = con.Write(bps)
			if err == nil {
				break
			}
			if i == 5 {
				agento.LogError("Error writing to influxdb: " + err.Error() + ", giving up")
			}
		}
	}
}

func reportHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "Only POST allowed", 400)
		return
	}

	var m agento.MachineStats
	d := json.NewDecoder(req.Body)
	d.UseNumber()
	err := d.Decode(&m)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	sendToInflux(m)
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
	err := config.LoadFromFile("/etc/agento.conf")
	agento.InitLogging(&config)

	if err != nil {
		agento.LogError("Configuration error: %s",
			err.Error())
		os.Exit(1)
	}

	http.HandleFunc("/report", reportHandler)
	http.HandleFunc("/health", healthHandler)

	addr := config.Server.Bind + ":" + strconv.Itoa(int(config.Server.Port))
	agento.LogInfo("agento server started, listening at " + addr)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		agento.LogError("ListenAndServe: %s", err.Error())
		log.Fatal("ListenAndServe: ", err)
	}

	agento.LogInfo("listening at %s", addr)
}
