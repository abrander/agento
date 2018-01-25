package configuration

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/abrander/agento/logger"
)

const (
	// StateDir is the path where Agento's state will be saves.
	StateDir = "/var/lib/agento/"
)

var (
	defaultConfig = `
[main]
includedir = "/etc/agento.d/"

[client]
enabled = false
interval = 1
secret = "insecure"

[server]
secret = "insecure"

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

[mongo]
enabled = false
url = "127.0.0.1"
database = "agento"
`

	// ProcPath is the path where Agento will expect the proc filesystem to be.
	// Will be /proc by default.
	ProcPath string

	// SysfsPath is the path where Agento expects to find the sys filesystem.
	// Default is /sys.
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

// InfluxdbConfiguration stores the configuration for the InfluxDB connection.
type InfluxdbConfiguration struct {
	URL             string `toml:"url"`
	Username        string `toml:"username"`
	Password        string `toml:"password"`
	Database        string `toml:"database"`
	RetentionPolicy string `toml:"retentionPolicy"`
	Retries         int    `toml:"retries"`
}

// ClientConfiguration stores the configuration for Agento as a client.
type ClientConfiguration struct {
	Enabled   bool   `toml:"enabled"`
	Interval  int    `toml:"interval"`
	Secret    string `toml:"secret"`
	ServerURL string `toml:"server-url"`
}

// HTTPConfiguration is the configuration for the built-in HTTP server.
type HTTPConfiguration struct {
	Enabled bool   `toml:"enabled"`
	Bind    string `toml:"bind"`
	Port    int16  `toml:"port"`
}

// HTTPSConfiguration is the configuration for the built-in HTTPS server.
type HTTPSConfiguration struct {
	Enabled  bool   `toml:"enabled"`
	Bind     string `toml:"bind"`
	Port     int16  `toml:"port"`
	KeyPath  string `toml:"key"`
	CertPath string `toml:"cert"`
}

// UDPConfiguration is the configuration for the UDP receiver.
type UDPConfiguration struct {
	Enabled  bool   `toml:"enabled"`
	Bind     string `toml:"bind"`
	Port     int16  `toml:"port"`
	Interval int    `toml:"interval"`
}

// ServerConfiguration stores the configuration for Agento as a server.
type ServerConfiguration struct {
	Influxdb InfluxdbConfiguration `toml:"influxdb"`
	HTTP     HTTPConfiguration     `toml:"http"`
	HTTPS    HTTPSConfiguration    `toml:"https"`
	Secret   string                `toml:"secret"`
	UDP      UDPConfiguration      `toml:"udp"`
}

// MongoConfiguration is the configuration for Agento's MongoDB client.
type MongoConfiguration struct {
	Enabled  bool   `toml:"enabled"`
	URL      string `toml:"url"`
	Database string `toml:"database"`
}

// MainConfiguration is the configuration for main behaviour of Agento.
type MainConfiguration struct {
	Includedir string `toml:"includedir"`
}

// Configuration is Agento's main configuration object.
type Configuration struct {
	Client   ClientConfiguration       `toml:"client"`
	Server   ServerConfiguration       `toml:"server"`
	Mongo    MongoConfiguration        `toml:"mongo"`
	Hosts    map[string]toml.Primitive `toml:"host"`
	Probes   map[string]toml.Primitive `toml:"probe"`
	Main     MainConfiguration         `toml:"agento"`
	metadata toml.MetaData
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// LoadFromFile loads configuration from TOML file at path.
func (c *Configuration) LoadFromFile(path string) error {
	c.LoadDefaults()

	// We default to agento.domain, try to guess it
	hostname, err := os.Hostname()

	if err != nil {
		return err
	}

	firstDot := strings.Index(hostname, ".")
	if firstDot > 0 {
		c.Client.ServerURL = "http://agento" + hostname[firstDot:] + ":12345/report"
	}

	// Read values from config file if it exists
	if fileExists(path) {
		c.metadata, err = toml.DecodeFile(path, &c)
		if err != nil {
			return err
		}
	}

	c.LoadFromEnvironment()

	if c.Client.Enabled && c.Client.ServerURL == "" {
		return errors.New("Could not determine server URL")
	}

	if c.Main.Includedir != "" {
		matches, err := filepath.Glob(c.Main.Includedir + "/*.conf")
		if err != nil {
			return err
		}

		if matches != nil {
			for _, match := range matches {
				c.metadata, err = toml.DecodeFile(match, &c)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// LoadDefaults sets the default configuration. Can later be overridden by
// LoadFromFile() and LoadFromEnvironment().
func (c *Configuration) LoadDefaults() {
	// Start by loading default values - should never err
	if _, err := toml.Decode(defaultConfig, &c); err != nil {
		panic(err.Error())
	}
}

// LoadFromEnvironment reads configuration from environment variables.
func (c *Configuration) LoadFromEnvironment() {
	envSecret := os.Getenv("AGENTO_SECRET")
	if envSecret != "" {
		c.Client.Secret = envSecret
		c.Server.Secret = envSecret
	}

	envServer := os.Getenv("AGENTO_SERVER_URL")
	if envServer != "" {
		c.Client.ServerURL = envServer
	}

	envInfluxdbURL := os.Getenv("AGENTO_INFLUXDB_URL")
	if envInfluxdbURL != "" {
		c.Server.Influxdb.URL = envInfluxdbURL
	}

	envMongoURL := os.Getenv("AGENTO_MONGO_URL")
	if envMongoURL != "" {
		c.Mongo.URL = envMongoURL
	}
}

// GetHostPrimitives will return enough for someone to decode [host.*] fields
// from the TOML file.
func (c *Configuration) GetHostPrimitives() map[string]toml.Primitive {
	return c.Hosts
}

// GetProbePrimitives will return enough for someone to decode [probe.*]
// fields from the TOML file.
func (c *Configuration) GetProbePrimitives() map[string]toml.Primitive {
	return c.Probes
}
