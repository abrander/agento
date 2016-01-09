package plugins

import (
	"reflect"
)

type Doc struct {
	Info struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"info"`
	Parameters   []Parameter       `json:"parameters"`
	Tags         map[string]string `json:"-"`
	Measurements map[string]string `json:"-"`
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
