package cpustats

import (
	"testing"

	"github.com/abrander/agento/plugins"
	"github.com/abrander/agento/plugins/transports/mock"
)

var (
	testData = []byte(`cpu  6038746 82650 1374615 237694432 351001 24 2586 0 0 0
cpu0 1562929 22772 382708 59311772 49120 13 390 0 0 0
cpu1 1598806 23329 388310 59296242 41929 6 179 0 0 0
cpu2 1488550 18503 297676 59399916 194084 0 410 0 0 0
cpu3 1388460 18044 305920 59686501 65867 4 1606 0 0 0
intr 305606156 24 0 0 0 0 0 0 0 1 3 0 0 0 0 0 0 33 0 0 13 0 0 0 454469 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 28 1493196 1516047 22 25812336 16659 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
ctxt 882885801
btime 1463676431
processes 98481
procs_running 2
procs_blocked 10
softirq 59361308 454349 29969552 124152 1952586 1224181 0 94964 15104956 421025 10015543
`)
)

func TestGather(t *testing.T) {
	transport := mocktransport.NewMock()
	agent := NewCpuStats().(plugins.Agent)

	err := agent.Gather(transport.(plugins.Transport))
	if err == nil {
		t.Fatal("did not catch error")
	}

	mock := transport.(*mocktransport.Mock)

	// An empty file should not generate errors.
	mock.SetFile("/proc/stat", []byte(""))

	err = agent.Gather(transport.(plugins.Transport))
	if err != nil {
		t.Fatal("empty (bogus) file generated an error")
	}

	mock.SetFile("/proc/stat", testData)
	err = agent.Gather(transport.(plugins.Transport))
	if err != nil {
		t.Errorf("Good file generated an error: %s", err.Error())
	}

	plugins.GenericAgentTest(t, agent)
}

func TestAgent(t *testing.T) {
	plugins.GenericAgentTest(t, NewCpuStats())
}
