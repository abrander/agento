package plugins

import (
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/influxdb/influxdb/client"
)

type Plugin interface {
	GetDoc() *Doc
}

type Parameter struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	EnumValues  []string `json:"enumValues"`
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

func getParams(elem reflect.Type) []Parameter {
	parameters := []Parameter{}

	if elem.Kind() != reflect.Struct {
		return parameters
	}

	l := elem.NumField()

	for i := 0; i < l; i++ {
		f := elem.Field(i)

		jsonName := f.Tag.Get("json")
		description := f.Tag.Get("description")

		if f.Anonymous {
			parameters = append(parameters, getParams(f.Type)...)
		} else if jsonName != "" && description != "" {
			p := Parameter{}

			p.Name = jsonName
			p.Type = f.Type.String()
			p.Description = description
			enum := f.Tag.Get("enum")
			if enum != "" {
				p.EnumValues = strings.Split(enum, ",")
				p.Type = "enum"
			} else {
				p.EnumValues = []string{}
			}

			parameters = append(parameters, p)
		}
	}

	return parameters
}
