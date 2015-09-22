package plugins

import (
	"log"
	"time"

	"github.com/influxdb/influxdb/client"

	"github.com/abrander/agento"
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
	start := time.Now()

	for name, p := range plugins {
		p.Gather()
		results[name] = p
	}

	results["g"] = time.Now().Sub(start)

	return results
}

type GatherDuration time.Duration

func (g *GatherDuration) Gather() error {
	return nil
}

func (g *GatherDuration) GetPoints() []client.Point {
	points := make([]client.Point, 1)

	points[0] = agento.SimplePoint("agento.GatherDuration", agento.Round(time.Duration(*g).Seconds()*1000.0, 1))

	return points
}
