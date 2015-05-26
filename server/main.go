package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/influxdb/influxdb/client"

	"agento"
)

var config = agento.Configuration{}

func sendToInflux(stats agento.MachineStats) {
	u, err := url.Parse(config.Server.Influxdb.Url)
	conf := client.Config{
		URL:      *u,
		Username: config.Server.Influxdb.Username,
		Password: config.Server.Influxdb.Password,
	}

	con, err := client.NewClient(conf)
	if err != nil {
		agento.LogError("InfluxDB error: %s", err.Error())
		log.Fatal(err)
	}

	m := stats.GetMap()

	points := make([]client.Point, len(m))

	i := 0

	for key, value := range m {
		points[i] = client.Point{
			Name: key,
			Fields: map[string]interface{}{
				"value": value,
			},
		}

		i++
	}

	bps := client.BatchPoints{
		Tags: map[string]string{
			"hostname": stats.Hostname,
		},
		Time:            time.Now(),
		Points:          points,
		Database:        config.Server.Influxdb.Database,
		RetentionPolicy: config.Server.Influxdb.RetentionPolicy,
	}

	_, err = con.Write(bps)
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

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), 400)
	}

	var m agento.MachineStats
	err = json.Unmarshal(body, &m)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	sendToInflux(m)
}

func main() {
	config.LoadDefaults()
	err := config.LoadFromFile("/etc/agento.json")
	agento.InitLogging(&config)

	if err != nil {
		agento.LogInfo("Could not read /etc/agento.json (%s). Using defaults",
			err.Error())
	}

	http.HandleFunc("/report", reportHandler)

	addr := config.Server.Bind + ":" + strconv.Itoa(int(config.Server.Port))
	agento.LogInfo("agento server started, listening at " + addr)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		agento.LogError("ListenAndServe: %s", err.Error())
		log.Fatal("ListenAndServe: ", err)
	}
}
