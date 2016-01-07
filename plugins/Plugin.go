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

type Doc struct {
	Info struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"info"`
	Parameters   []Parameter       `json:"parameters"`
	Tags         map[string]string `json:"-"`
	Measurements map[string]string `json:"-"`
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

func GetDoc() map[string]*Doc {
	docs := make(map[string]*Doc)

	for shortName, p := range plugins {
		agent, ok := p.(Plugin)
		if ok {
			docs[shortName] = agent.GetDoc()
		}
	}

	return docs
}

func NewDoc(description string) *Doc {
	var doc Doc

	doc.Info.Description = description
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

func getDescriptions(m map[string]PluginConstructor) map[string]*Doc {
	r := make(map[string]*Doc)
	for name, c := range m {
		doc := c().(Plugin).GetDoc()

		if doc.Info.Name == "" {
			doc.Info.Name = name
		}

		elem := reflect.TypeOf(c().(Plugin)).Elem()
		doc.Parameters = getParams(elem)

		r[name] = doc
	}

	return r
}

func AvailableAgents() map[string]*Doc {
	p := getPlugins(reflect.TypeOf((*Agent)(nil)).Elem())
	return getDescriptions(p)
}

func AvailablePlugins() map[string]*Doc {
	p := getPlugins(reflect.TypeOf((*Plugin)(nil)).Elem())
	return getDescriptions(p)
}

func AvailableTransports() map[string]*Doc {
	p := getPlugins(reflect.TypeOf((*Transport)(nil)).Elem())
	return getDescriptions(p)
}
