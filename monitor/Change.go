package monitor

import (
	"sync"
)

type (
	Change struct {
		Type    string      `json:"type"`
		Payload interface{} `json:"payload"`
	}
)

var (
	changesLock sync.Mutex
	changes     []chan Change
)

func SubscribeChanges() chan Change {
	channel := make(chan Change)
	changesLock.Lock()
	changes = append(changes, channel)
	changesLock.Unlock()

	return channel
}

func UnsubscribeChanges(ch chan Change) {
	changesLock.Lock()

	for i, c := range changes {
		if ch == c {
			changes = append(changes[:i], changes[i+1:]...)
			break
		}
	}
	changesLock.Unlock()
}

func broadcastChange(typ string, payload interface{}) {
	change := Change{
		Type:    typ,
		Payload: payload,
	}

	changesLock.Lock()
	for _, ch := range changes {
		ch <- change
	}
	changesLock.Unlock()
}
