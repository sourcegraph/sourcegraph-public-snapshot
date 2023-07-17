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
	if MockGetRegisteredEventTypes != nil {
		return MockGetRegisteredEventTypes()
	}

	registeredEventTypes.RLock()
	defer registeredEventTypes.RUnlock()

	return append(types, registeredEventTypes.types...)
}

var MockGetRegisteredEventTypes func() []EventType

// RegisterEventType registers a new outbound webhook event type, thereby making
// it available in the webhook admin UI.
//
// It is generally expected that this will be invoked from init() functions. It MUST NOT
// be invoked before init().
//
// For an example of how to register events, see
// internal/batches/webhooks/event_types.go.
func RegisterEventType(eventType EventType) {
	registeredEventTypes.Lock()
	defer registeredEventTypes.Unlock()

	registeredEventTypes.types = append(registeredEventTypes.types, eventType)
}
