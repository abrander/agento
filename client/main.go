package main

import (
	"bytes"
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

	full_frequency := config.Client.Interval
	i := full_frequency

	machineStats := agento.MachineStats{}
	receiver := agento.MachineStats{}
	for {
		machineStats.Gather()
		json, err := machineStats.GetJson(i%full_frequency != 0)

		if err == nil {
			receiver.ReadJson(json)
			res, err := http.Post(config.Client.ServerUrl, "image/jpeg", bytes.NewReader(json))
			if err != nil {
				fmt.Println(err, i)
				continue
			}
			io.Copy(ioutil.Discard, res.Body)
			res.Body.Close()
		} else {
			fmt.Println(err, i)
		}

		i++

		time.Sleep(time.Second * time.Duration(config.Client.Interval))
	}

	return
}
