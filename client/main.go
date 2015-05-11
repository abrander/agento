package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"agento"
)

func main() {
	config := agento.Configuration{}
	config.LoadDefaults()
	config.LoadFromFile("/etc/agento.json")

	machineStats := agento.MachineStats{}
	c := time.Tick(time.Second * time.Duration(config.Client.Interval))
	for _ = range c {
		machineStats.Gather()
		json, err := json.Marshal(machineStats)

		if err == nil {
			res, err := http.Post(config.Client.ServerUrl, "image/jpeg", bytes.NewReader(json))
			if err != nil {
				fmt.Println(err)
				continue
			}
			io.Copy(ioutil.Discard, res.Body)
			res.Body.Close()
		} else {
			fmt.Println(err)
		}

	}

	return
}
