package monitor

import (
	"sync"

	"github.com/abrander/agento/userdb"
)

type (
	Emitter interface {
		Subscribe(subject userdb.Subject) chan Change
		Unsubscribe(chan Change)
	}

	Broadcaster interface {
		Broadcast(typ string, payload userdb.Object)
	}

	SimpleEmitter struct {
		lock      sync.Mutex
		listeners []*Listener
	}

	Listener struct {
		subject userdb.Subject
		channel chan Change
	}

	Change struct {
		Type    string        `json:"type"`
		Payload userdb.Object `json:"payload"`
	}
)

func NewSimpleEmitter() *SimpleEmitter {
	return &SimpleEmitter{}
}

func (s *SimpleEmitter) Subscribe(subject userdb.Subject) chan Change {
	listener := &Listener{
		subject: subject,
		channel: make(chan Change),
	}

	s.lock.Lock()
	s.listeners = append(s.listeners, listener)
	s.lock.Unlock()

	return listener.channel
}

func (s *SimpleEmitter) Unsubscribe(ch chan Change) {
	s.lock.Lock()

	for i, l := range s.listeners {
		if ch == l.channel {
			s.listeners = append(s.listeners[:i], s.listeners[i+1:]...)
			break
		}
	}
	s.lock.Unlock()
}

func (s *SimpleEmitter) Broadcast(typ string, payload userdb.Object) {
	change := Change{
		Type:    typ,
		Payload: payload,
	}

	s.lock.Lock()
	for _, listener := range s.listeners {
		if listener.subject.CanAccess(payload) == nil {
			listener.channel <- change
		}
	}
	s.lock.Unlock()
}
