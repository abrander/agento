package plugins

import (
	"errors"
	"testing"

	"github.com/influxdata/influxdb/client/v2"
)

type (
	// Agent describes the interface an agent must implement. An agent is a
	// collector collecting data and/or metrics from an underlying source.
	Agent interface {
		Gather(transport Transport) error
		GetPoints() []*client.Point
	}
)

// GetAgent will return an agent of type id or nil plus an error if the
// agent was not found.
func GetAgent(id string) (Agent, error) {
	// Try to find a constructor.
	c, found := pluginConstructors[id]
	if !found {
		return nil, errors.New("Agent " + id + " not found")
	}

	// Instantiate.
	plugin := c()

	// Check if the plugin is in fact an agent.
	agent, ok := plugin.(Agent)
	if !ok {
		return nil, errors.New("Agent " + id + " not found")
	}

	return agent, nil
}

// GenericAgentTest can be called from local agent _test files to test agents
// for conformance.
func GenericAgentTest(t *testing.T, i interface{}) {
	agent, ok := i.(Agent)
	if !ok {
		t.Errorf("Agent %T does not implement the Agent interface", i)
		return
	}

	plugin, ok := i.(Plugin)
	if !ok {
		t.Errorf("Agent %T does not implement the Plugin interface", i)
		return
	}

	doc := plugin.GetDoc()
	points := agent.GetPoints()

	for _, point := range points {
		name := point.Name()

		// Check measurement documentation.
		_, found := doc.Measurements[name]
		if !found {
			t.Errorf("Measurement '%s' not documented on %T", name, i)
			doc.Measurements[name] = "do not complain again"
		}

		// Check tags.
		for tag := range point.Tags() {
			_, found = doc.Tags[tag]
			if !found {
				t.Errorf("Tag '%s' not documented on %T", tag, i)
				doc.Tags[tag] = "do not complain again"
			}
		}
	}
}
