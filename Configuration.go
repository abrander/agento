package agento

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strings"
)

var defaultConfig = []byte(`{
	"client": {
		"interval": 1
	},
	"server": {
		"port": 12345,
		"influxdb": {
			"url": "http://localhost:8086/",
			"username": "root",
			"password": "root",
			"database": "agento",
			"retentionPolicy": "default"
		}
	}
}
`)

type InfluxdbConfiguration struct {
	Url             string `json:"url"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	Database        string `json:"database"`
	RetentionPolicy string `json:"retentionPolicy"`
}

type ClientConfiguration struct {
	Interval  int    `json:"interval"`
	ServerUrl string `json:"serverUrl"`
}

type ServerConfiguration struct {
	Influxdb InfluxdbConfiguration `json:"influxdb"`
	Port     int16                 `json:"port"`
}

type Configuration struct {
	Client ClientConfiguration `json:"client"`
	Server ServerConfiguration `json:"server"`
}

func (c *Configuration) LoadDefaults() error {
	err := json.Unmarshal(defaultConfig, c)
	if err != nil {
		return err
	}

	hostname, err := os.Hostname()

	if err != nil {
		return err
	}

	firstDot := strings.Index(hostname, ".")
	if firstDot <= 0 {
		return errors.New("Could not extract domain name from '" + hostname + "'")
	}

	c.Client.ServerUrl = "http://agento" + hostname[firstDot:] + ":12345/report"

	return nil
}

func (c *Configuration) LoadFromFile(path string) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(content, c)
	if err != nil {
		return err
	}

	return nil
}
