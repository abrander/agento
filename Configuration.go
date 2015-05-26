package agento

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

var defaultConfig = `
[client]
interval = 1

[server]
bind = "0.0.0.0"
port = 12345

[server.influxdb]
url = "http://localhost:8086/"
username = "root"
password = "root"
database = "agento"
retentionPolicy = "default"
`

type InfluxdbConfiguration struct {
	Url             string `toml:"url"`
	Username        string `toml:"username"`
	Password        string `toml:"password"`
	Database        string `toml:"database"`
	RetentionPolicy string `toml:"retentionPolicy"`
}

type ClientConfiguration struct {
	Interval  int    `toml:"interval"`
	ServerUrl string `toml:"serverUrl"`
}

type ServerConfiguration struct {
	Influxdb InfluxdbConfiguration `toml:"influxdb"`
	Bind     string                `toml:"bind"`
	Port     int16                 `toml:"port"`
}

type Configuration struct {
	Client ClientConfiguration `toml:"client"`
	Server ServerConfiguration `toml:"server"`
}

func (c *Configuration) LoadDefaults() error {
	if _, err := toml.Decode(defaultConfig, &c); err != nil {
		return err
	}

	hostname, err := os.Hostname()

	if err != nil {
		return err
	}

	firstDot := strings.Index(hostname, ".")
	if firstDot <= 0 {
		c.Client.ServerUrl = "http://agento.example.com:12345/report"

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
