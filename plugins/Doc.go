package plugins

import (
	"reflect"
)

// Doc represents end user documentation for a plugin.
type Doc struct {
	Info struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"info"`
	Parameters   []Parameter       `json:"parameters"`
	Tags         map[string]string `json:"-"`
	Measurements map[string]string `json:"-"`
}

// NewDoc will instantiate a new Doc. Can be used from plugins to build GetDoc().
func NewDoc(description string) *Doc {
	var doc Doc

	doc.Info.Description = description
	doc.Measurements = make(map[string]string)
	doc.Tags = make(map[string]string)

	return &doc
}

// AddMeasurement will add documentation for a measurement.
func (d *Doc) AddMeasurement(key string, description string, unit string) {
	d.Measurements[key] = description + " (" + unit + ")"
}

// AddTag will add documentation for a tag.
func (d *Doc) AddTag(key string, description string) {
	d.Tags[key] = description
}

// GetDoc will return a map of documentation for all plugins.
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

// GetDocAgents behaves like GetDoc(), but will only return documentaiton for
// agents.
func GetDocAgents() map[string]*Doc {
	p := getPlugins(reflect.TypeOf((*Agent)(nil)).Elem())
	return getDescriptions(p)
}

// GetDocTransports behaves like GetDoc(), but will only return documentaiton
// for transports.
func GetDocTransports() map[string]*Doc {
	p := getPlugins(reflect.TypeOf((*Transport)(nil)).Elem())
	return getDescriptions(p)
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
