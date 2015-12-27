package plugins

import (
	"log"
	"reflect"
	"time"

	"github.com/influxdb/influxdb/client"
)

type Plugin interface {
}

type Doc struct {
	Description  string
	Tags         map[string]string
	Measurements map[string]string
}

type PluginConstructor func() Plugin

var plugins = map[string]Plugin{}
var pluginConstructors = map[string]func() Plugin{}

func Register(shortName string, constructor PluginConstructor) {
	_, exists := plugins[shortName]
	if exists {
		log.Fatalf("plugins.Register(): Duplicate shortname: '%s' (%T and %T)\n", shortName, plugins[shortName], constructor())
		return
	}

	plugins[shortName] = constructor()
	pluginConstructors[shortName] = constructor
}

func GatherAll() Results {
	transport := GetPlugin("localtransport").(Transport)

	var results = Results{}
	start := time.Now()

	for name, p := range plugins {
		agent, ok := p.(Agent)
		if ok {
			agent.Gather(transport)
			results[name] = p
		}
	}

	results["g"] = GatherDuration(time.Now().Sub(start))

	return results
}

func GetDoc() map[string]*Doc {
	docs := make(map[string]*Doc)

	for shortName, p := range plugins {
		agent, ok := p.(Agent)
		if ok {
			docs[shortName] = agent.GetDoc()
		}
	}

	return docs
}

func NewDoc(description string) *Doc {
	var doc Doc

	doc.Description = description
	doc.Measurements = make(map[string]string)
	doc.Tags = make(map[string]string)

	return &doc
}

func (d *Doc) AddMeasurement(key string, description string, unit string) {
	d.Measurements[key] = description + " (" + unit + ")"
}

func (d *Doc) AddTag(key string, description string) {
	d.Tags[key] = description
}

type GatherDuration time.Duration

func (g *GatherDuration) Gather() error {
	return nil
}

func (g *GatherDuration) GetPoints() []client.Point {
	points := make([]client.Point, 1)

	points[0] = SimplePoint("agento.GatherDuration", Round(time.Duration(*g).Seconds()*1000.0, 1))

	return points
}

func getPlugins(iType reflect.Type) map[string]PluginConstructor {
	r := make(map[string]PluginConstructor)

	for name, plugin := range plugins {
		pType := reflect.TypeOf(plugin)
		//		elem := reflect.TypeOf(plugin).Elem()
		if pType.Implements(iType) {
			r[name] = pluginConstructors[name]
		}
	}

	return r
}

func GetPlugin(name string) Plugin {
	return pluginConstructors[name]()
}

func GetPlugins() map[string]PluginConstructor {
	return getPlugins(reflect.TypeOf((*Plugin)(nil)).Elem())
}

func GetAgents() map[string]PluginConstructor {
	return getPlugins(reflect.TypeOf((*Agent)(nil)).Elem())
}

func GetTransports() map[string]PluginConstructor {
	return getPlugins(reflect.TypeOf((*Transport)(nil)).Elem())
}
