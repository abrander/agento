package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"agento"
)

func main() {
	config := agento.Configuration{}
	config.LoadDefaults()
	err := config.LoadFromFile("/etc/agento.json")
	agento.InitLogging(&config)

	if err != nil {
		agento.LogInfo("Could not read /etc/agento.json (%s). Using defaults and logging to %s",
			err.Error(),
			config.Client.ServerUrl)
	}

	machineStats := agento.MachineStats{}

	// We need to gather one unreported set of metrics. It's needed for
	// calculating deltas on first real report
	machineStats.Gather()

	c := time.Tick(time.Second * time.Duration(config.Client.Interval))
	for _ = range c {
		machineStats.Gather()
		json, err := json.Marshal(machineStats)

		if err == nil {
			res, err := http.Post(config.Client.ServerUrl, "image/jpeg", bytes.NewReader(json))
			if err != nil {
				agento.LogError(err.Error())
				continue
			}
			io.Copy(ioutil.Discard, res.Body)
			res.Body.Close()
		} else {
			agento.LogError(err.Error())
		}

	}

	return
}
