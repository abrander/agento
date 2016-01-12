package plugins

import (
	"log"
	"reflect"
	"strings"
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
