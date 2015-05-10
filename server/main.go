package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/influxdb/influxdb/client"

	"agento"
)

var config = agento.Configuration{}

func echoHandler(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, req.RequestURI)
}

func sendToInflux(stats agento.MachineStats) {
	u, err := url.Parse(config.Server.Influxdb.Url)
	conf := client.Config{
		URL:      *u,
		Username: config.Server.Influxdb.Username,
		Password: config.Server.Influxdb.Password,
	}

	con, err := client.NewClient(conf)
	if err != nil {
		log.Fatal(err)
	}

	_, _, err = con.Ping()
	if err != nil {
		log.Fatal(err)
	}

	memoryMap := stats.MemoryStats.GetMap()
	cpuMap := stats.CpuStats.GetMap()
	diskMap := stats.DiskStats.GetMap()
	netMap := stats.NetStats.GetMap()
	loadMap := stats.LoadStats.GetMap()

	points := make([]client.Point, len(*memoryMap)+
		len(*cpuMap)+
		len(*diskMap)+
		len(*netMap)+
		len(*loadMap))

	i := 0

	for key, value := range *memoryMap {
		points[i] = client.Point{
			Name: key,
			Fields: map[string]interface{}{
				"value": value,
			},
		}

		i++
	}

	for key, value := range *cpuMap {
		points[i] = client.Point{
			Name: key,
			Fields: map[string]interface{}{
				"value": value,
			},
		}

		i++
	}

	for key, value := range *diskMap {
		points[i] = client.Point{
			Name: key,
			Fields: map[string]interface{}{
				"value": value,
			},
		}

		i++
	}

	for key, value := range *netMap {
		points[i] = client.Point{
			Name: key,
			Fields: map[string]interface{}{
				"value": value,
			},
		}

		i++
	}

	for key, value := range *loadMap {
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
		Timestamp:       time.Now(),
		Points:          points,
		Database:        config.Server.Influxdb.Database,
		RetentionPolicy: config.Server.Influxdb.RetentionPolicy,
	}

	_, err = con.Write(bps)
	if err != nil {
		log.Fatal(err)
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

	// Print some debug info
	//	json, _ := m.GetJson(false)
	//	os.Stdout.Write(json)
	//	fmt.Println("")
	sendToInflux(m)
}

func main() {
	config.LoadDefaults()
	config.LoadFromFile("/etc/agento.json")
	http.HandleFunc("/echo/", echoHandler)
	http.HandleFunc("/report", reportHandler)
	err := http.ListenAndServe(":12345", nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
