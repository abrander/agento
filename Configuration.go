package agento

import (
	"errors"
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
	ServerUrl string `toml:"server-url"`
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

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func (c *Configuration) LoadFromFile(path string) error {
	// Start by loading default values - should never err
	if _, err := toml.Decode(defaultConfig, &c); err != nil {
		return err
	}

	// We default to agento.domain, try to guess it
	hostname, err := os.Hostname()

	if err != nil {
		return err
	}

	firstDot := strings.Index(hostname, ".")
	if firstDot > 0 {
		c.Client.ServerUrl = "http://agento" + hostname[firstDot:] + ":12345/report"
	}

	// Read values from config file if it exists
	if fileExists(path) {
		if _, err := toml.DecodeFile(path, &c); err != nil {
			return err
		}
	}

	if c.Client.ServerUrl == "" {
		return errors.New("Could not determine server URL")
	}

	return nil
}
