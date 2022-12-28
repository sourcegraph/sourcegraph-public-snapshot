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

func RegisterEventType(eventType EventType) {
	registeredEventTypes.Lock()
	defer registeredEventTypes.Unlock()

	registeredEventTypes.types = append(registeredEventTypes.types, eventType)
}
