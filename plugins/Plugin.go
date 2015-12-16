package plugins

import (
	"log"
	"time"

	"github.com/influxdb/influxdb/client"
)

type Plugin interface {
	Gather() error
	GetPoints() []client.Point
	GetDoc() *Doc
}

type Doc struct {
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
	var results = Results{}
	start := time.Now()

	for name, p := range plugins {
		p.Gather()
		results[name] = p
	}

	results["g"] = GatherDuration(time.Now().Sub(start))

	return results
}

func GetDoc() map[string]*Doc {
	docs := make(map[string]*Doc)

	for shortName, p := range plugins {
		docs[shortName] = p.GetDoc()
	}

	return docs
}

func NewDoc() *Doc {
	var doc Doc

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
