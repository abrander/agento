package plugins

import (
	"log"
	"reflect"
	"strings"
)

// Plugin is a basic interface all plugins must implement.
type Plugin interface {
	GetDoc() *Doc
}

// Parameter describes the user supplied parameters of a plugin (most often
// an agent).
type Parameter struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	EnumValues  []string `json:"enumValues"`
}

// PluginConstructor is the type for a function that will instantiate a plugin.
type PluginConstructor func() Plugin

// A map al already instantiated and configured plugins.
var plugins = map[string]Plugin{}

// Constructors for all plugins. Please note that the plugin will *not* be
// configured after calling the constructor.
var pluginConstructors = map[string]func() Plugin{}

// Register will register (at runtime) a new plugin. This should be done from
// init() in the plugin.
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

// GetPlugins will return a list of constructors for all compiled plugins.
func GetPlugins() map[string]PluginConstructor {
	return getPlugins(reflect.TypeOf((*Plugin)(nil)).Elem())
}

// GetAgents will return a list of constructors for all compiled agents.
func GetAgents() map[string]PluginConstructor {
	return getPlugins(reflect.TypeOf((*Agent)(nil)).Elem())
}

// GetTransports will return a list of constructors for all compiled transports.
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
