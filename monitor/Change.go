package monitor

import (
	"sync"
)

type (
	Emitter interface {
		Subscribe() chan Change
		Unsubscribe(chan Change)
	}

	Broadcaster interface {
		Broadcast(typ string, payload interface{})
	}

	SimpleEmitter struct {
		changesLock sync.Mutex
		changes     []chan Change
	}

	Change struct {
		Type    string      `json:"type"`
		Payload interface{} `json:"payload"`
	}
)

func NewSimpleEmitter() *SimpleEmitter {
	return &SimpleEmitter{}
}

func (s *SimpleEmitter) Subscribe() chan Change {
	channel := make(chan Change)
	s.changesLock.Lock()
	s.changes = append(s.changes, channel)
	s.changesLock.Unlock()

	return channel
}

func (s *SimpleEmitter) Unsubscribe(ch chan Change) {
	s.changesLock.Lock()

	for i, c := range s.changes {
		if ch == c {
			s.changes = append(s.changes[:i], s.changes[i+1:]...)
			break
		}
	}
	s.changesLock.Unlock()
}

func (s *SimpleEmitter) Broadcast(typ string, payload interface{}) {
	change := Change{
		Type:    typ,
		Payload: payload,
	}

	s.changesLock.Lock()
	for _, ch := range s.changes {
		ch <- change
	}
	s.changesLock.Unlock()
}
