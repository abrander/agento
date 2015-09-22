package plugins

import (
	"log"

	"github.com/influxdb/influxdb/client"
)

type Plugin interface {
	Gather() error
	GetPoints() []client.Point
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
	var results = Results{}

	for name, p := range plugins {
		p.Gather()
		results[name] = p
	}

	return results
}
