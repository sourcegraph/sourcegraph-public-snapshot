package outbound

import "sync"

type EventType struct {
	Key         string
	Description string
}

type eventTypes struct {
	sync.RWMutex
	types []EventType
}

var registeredEventTypes = eventTypes{types: []EventType{}}

func GetRegisteredEventTypes() (types []EventType) {
	registeredEventTypes.RLock()
	defer registeredEventTypes.RUnlock()

	return append(types, registeredEventTypes.types...)
}

// RegisterEventType registers a new outbound webhook event type, thereby making
// it available in the webhook admin UI.
//
// It is generally expected that this will be invoked from init() functions. It
// MUST NOT be invoked before init().
func RegisterEventType(eventType EventType) {
	registeredEventTypes.Lock()
	defer registeredEventTypes.Unlock()

	registeredEventTypes.types = append(registeredEventTypes.types, eventType)
}
