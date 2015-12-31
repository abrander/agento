package client

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/abrander/agento/configuration"
	"github.com/abrander/agento/logger"
	"github.com/abrander/agento/plugins"
)

func GatherAndReport(clientConfig configuration.ClientConfiguration) {
	logger.Yellow("client", "agento client started, reporting to %s", clientConfig.ServerUrl)

	// Randomize our start time to avoid a big cluster reporting at the exact same time
	time.Sleep(time.Second * time.Duration(rand.Intn(clientConfig.Interval)))

	c := time.Tick(time.Second * time.Duration(clientConfig.Interval))
	for _ = range c {
		results := plugins.GatherAll()
		json, err := json.Marshal(results)

		if err == nil {
			client := &http.Client{}
			req, err := http.NewRequest("POST", clientConfig.ServerUrl, bytes.NewReader(json))
			if err != nil {
				logger.Error("client", "%s", err.Error())
				continue
			}

			if clientConfig.Secret != "" {
				req.Header.Add("X-Agento-Secret", clientConfig.Secret)
			}

			res, err := client.Do(req)
			if err != nil {
				logger.Error("client", "%s", err.Error())
				continue
			}
			io.Copy(ioutil.Discard, res.Body)
			res.Body.Close()
		} else {
			logger.Error("client", "%s", err.Error())
		}

	}
}