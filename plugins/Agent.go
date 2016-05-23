package plugins

import (
	"errors"

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
