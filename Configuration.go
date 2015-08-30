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
secret = ""

[server.http]
enabled = true
bind = "0.0.0.0"
port = 12345

[server.https]
enabled = false
bind = "0.0.0.0"
port = 443
key = "/etc/agento/ssl.key"
cert = "/etc/agento/ssl.cert"

[server.influxdb]
url = "http://localhost:8086/"
username = "root"
password = "root"
database = "agento"
retentionPolicy = "default"
retries = 0
`

type InfluxdbConfiguration struct {
	Url             string `toml:"url"`
	Username        string `toml:"username"`
	Password        string `toml:"password"`
	Database        string `toml:"database"`
	RetentionPolicy string `toml:"retentionPolicy"`
	Retries         int    `toml:"retries"`
}

type ClientConfiguration struct {
	Interval  int    `toml:"interval"`
	Secret    string `toml:"secret"`
	ServerUrl string `toml:"server-url"`
}

type HttpConfiguration struct {
	Enabled bool   `toml:"enabled"`
	Bind    string `toml:"bind"`
	Port    int16  `toml:"port"`
}

type HttpsConfiguration struct {
	Enabled  bool   `toml:"enabled"`
	Bind     string `toml:"bind"`
	Port     int16  `toml:"port"`
	KeyPath  string `toml:"key"`
	CertPath string `toml:"cert"`
}

type ServerConfiguration struct {
	Influxdb InfluxdbConfiguration `toml:"influxdb"`
	Http     HttpConfiguration     `toml:"http"`
	Https    HttpsConfiguration    `toml:"https"`
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
