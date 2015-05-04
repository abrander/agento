package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"agento"
)

func main() {
	hostname, err := os.Hostname()

	if err != nil {
		fmt.Println("Could not determine local hostname: " + err.Error())
		return
	}

	firstDot := strings.Index(hostname, ".")
	if firstDot <= 0 {
		fmt.Println("Could not extract domain name from '" + hostname + "'")
		return
	}

	targetServer := "agento" + hostname[firstDot:]
	full_frequency := 1
	i := full_frequency

	machineStats := agento.MachineStats{}
	receiver := agento.MachineStats{}
	for {
		machineStats.Gather()
		json, err := machineStats.GetJson(i%full_frequency != 0)

		if err == nil {
			receiver.ReadJson(json)
			res, err := http.Post("http://"+targetServer+":12345/report", "image/jpeg", bytes.NewReader(json))
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

		time.Sleep(time.Second * 1)
	}

	return
}
