package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"

	"agento"
)

func main() {
	config := agento.Configuration{}
	err := config.LoadFromFile("/etc/agento.conf")
	agento.InitLogging(&config)

	if err != nil {
		agento.LogError("Configuration error: %s",
			err.Error())
		os.Exit(1)
	}

	agento.LogInfo("agento client started, reporting to %s", config.Client.ServerUrl)

	machineStats := agento.MachineStats{}

	// Randomize our start time to avoid a big cluster reporting at the exact same time
	time.Sleep(time.Second * time.Duration(rand.Intn(config.Client.Interval)))

	// We need to gather one unreported set of metrics. It's needed for
	// calculating deltas on first real report
	machineStats.Gather()

	c := time.Tick(time.Second * time.Duration(config.Client.Interval))
	for _ = range c {
		machineStats.Gather()
		json, err := json.Marshal(machineStats)

		if err == nil {
			client := &http.Client{}
			req, err := http.NewRequest("POST", config.Client.ServerUrl, bytes.NewReader(json))
			if err != nil {
				agento.LogError(err.Error())
				continue
			}

			if config.Client.Secret != "" {
				req.Header.Add("X-Agento-Secret", config.Client.Secret)
			}

			res, err := client.Do(req)
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
