package linuxhost

import (
	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/timeseries"
)

type LinuxHost struct {
	Agents map[string]plugins.Agent `json:"agents" bson:"-"`
}

var agentIds = []string{
	"cpustats",
	"diskio",
	"entropy",
	"hostname",
	"load",
	"memory",
	"netfilter",
	"netio",
	"netstat",
	"openfiles",
	"sockets",
}

func init() {
	plugins.Register("linuxhost", NewLinuxHost)
}

func NewLinuxHost() interface{} {
	return new(LinuxHost)
}

func (l *LinuxHost) Gather(transport plugins.Transport) error {
	agents := plugins.GetAgents()
	l.Agents = make(map[string]plugins.Agent)

	for _, agentId := range agentIds {
		agent, found := agents[agentId]

		if found {
			l.Agents[agentId] = agent().(plugins.Agent)
			err := l.Agents[agentId].Gather(transport)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (l *LinuxHost) GetPoints() []*timeseries.Point {
	points := make([]*timeseries.Point, 0, 300)

	for _, p := range l.Agents {
		agent, ok := p.(plugins.Agent)
		if ok {
			points = append(points, agent.GetPoints()...)
		}
	}

	return points
}

func (l *LinuxHost) GetDoc() *plugins.Doc {
	doc := plugins.NewDoc("Linux host")

	return doc
}

// Ensure compliance
var _ plugins.Agent = (*LinuxHost)(nil)
