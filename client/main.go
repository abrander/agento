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

	// We need to gather one unreported set of metrics. It's needed for
	// calculating deltas on first real report
	machineStats.Gather()

	// Randomize our start time to avoid a big cluster reporting at the exact same time
	time.Sleep(time.Second * time.Duration(rand.Intn(config.Client.Interval)))

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
