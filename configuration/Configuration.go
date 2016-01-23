package configuration

import (
	"errors"
	"os"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/abrander/agento/logger"
)

const (
	StateDir = "/var/lib/agento/"
)

var (
	defaultConfig = `
[client]
enabled = true
interval = 1
secret = "insecure"

[server]
secret = ""

[server.http]
enabled = false
bind = "0.0.0.0"
port = 12345

[server.https]
enabled = false
bind = "0.0.0.0"
port = 443
key = "/etc/agento/ssl.key"
cert = "/etc/agento/ssl.cert"

[server.udp]
enabled = false
bind = "0.0.0.0"
port = 12345
interval = 60

[server.influxdb]
url = "http://localhost:8086/"
username = "root"
password = "root"
database = "agento"
retentionPolicy = "default"
retries = 0

[monitor]
enabled = false

[monitor.mongo]
url = "127.0.0.1"
database = "agento"
`

	ProcPath  string
	SysfsPath string
)

func init() {
	ProcPath = os.Getenv("AGENTO_PROC_PATH")
	SysfsPath = os.Getenv("AGENTO_SYSFS_PATH")

	if ProcPath == "" {
		ProcPath = "/proc"
	}

	if SysfsPath == "" {
		SysfsPath = "/sys"
	}

	_, err := os.Stat(StateDir)
	if err != nil {
		uid := os.Getuid()
		gid := os.Getgid()
		logger.Error("config", "Please run:\nsudo mkdir -p %s && sudo chown %d.%d %s\n", StateDir, uid, gid, StateDir)
	}
}

type InfluxdbConfiguration struct {
	Url             string `toml:"url"`
	Username        string `toml:"username"`
	Password        string `toml:"password"`
	Database        string `toml:"database"`
	RetentionPolicy string `toml:"retentionPolicy"`
	Retries         int    `toml:"retries"`
}

type ClientConfiguration struct {
	Enabled   bool   `toml:"enabled"`
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

type UdpConfiguration struct {
	Enabled  bool   `toml:"enabled"`
	Bind     string `toml:"bind"`
	Port     int16  `toml:"port"`
	Interval int    `toml:"interval"`
}

type ServerConfiguration struct {
	Influxdb InfluxdbConfiguration `toml:"influxdb"`
	Http     HttpConfiguration     `toml:"http"`
	Https    HttpsConfiguration    `toml:"https"`
	Secret   string                `toml:"secret"`
	Udp      UdpConfiguration      `toml:"udp"`
}

type MonitorConfiguration struct {
	Enabled bool               `toml:"enabled"`
	Mongo   MongoConfiguration `toml:mongo`
}

type MongoConfiguration struct {
	Url      string `toml:"url`
	Database string `toml:database`
}

type Configuration struct {
	Client  ClientConfiguration  `toml:"client"`
	Server  ServerConfiguration  `toml:"server"`
	Monitor MonitorConfiguration `toml:"monitor"`
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
	c.LoadDefaults()

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

	c.LoadFromEnvironment()

	if c.Client.ServerUrl == "" {
		return errors.New("Could not determine server URL")
	}

	return nil
}

func (c *Configuration) LoadDefaults() {
	// Start by loading default values - should never err
	if _, err := toml.Decode(defaultConfig, &c); err != nil {
		panic(err.Error())
	}
}

func (c *Configuration) LoadFromEnvironment() {
	envSecret := os.Getenv("AGENTO_SECRET")
	if envSecret != "" {
		c.Client.Secret = envSecret
		c.Server.Secret = envSecret
	}

	envServer := os.Getenv("AGENTO_SERVER_URL")
	if envServer != "" {
		c.Client.ServerUrl = envServer
	}

	envInfluxdbUrl := os.Getenv("AGENTO_INFLUXDB_URL")
	if envInfluxdbUrl != "" {
		c.Server.Influxdb.Url = envInfluxdbUrl
	}

	envMongoUrl := os.Getenv("AGENTO_MONGO_URL")
	if envMongoUrl != "" {
		c.Monitor.Mongo.Url = envMongoUrl
	}
}
