package netstat

import (
	"testing"

	"github.com/abrander/agento/plugins"
)

func TestAgent(t *testing.T) {
	plugins.GenericAgentTest(t, NewNetStats())
}
