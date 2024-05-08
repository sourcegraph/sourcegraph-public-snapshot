package honey

import (
	"go.opentelemetry.io/otel/attribute"
	"sync"
)

// NonSendingEvent returns an Event implementation that does not actually send data to
// Honeycomb, but instead aggregates it in memory. This is useful for testing and
// logging purposes.
func NonSendingEvent() Event {
	return &nonSendingEvent{}
}

// nonSendingEvent returns an Event implementation that does not actually send data to
// Honeycomb, but instead aggregates it in memory. This is useful for testing and
// logging purposes.
type nonSendingEvent struct {
	fieldsWritingMu sync.RWMutex
	fields          map[string]any
}

// Dataset returns the destination dataset of this event.
//
// For nonSendingEvent, this is always an empty string.
func (e *nonSendingEvent) Dataset() string {
	return ""
}

// AddField adds a single key-value pair to this event.
func (e *nonSendingEvent) AddField(key string, val any) {
	e.fieldsWritingMu.Lock()
	defer e.fieldsWritingMu.Unlock()

	if e.fields == nil {
		e.fields = make(map[string]any)
	}

	e.fields[key] = val
}

// AddAttributes adds each otel/attribute key-value field to this event.
func (e *nonSendingEvent) AddAttributes(values []attribute.KeyValue) {
	e.fieldsWritingMu.Lock()
	defer e.fieldsWritingMu.Unlock()

	if e.fields == nil {
		e.fields = make(map[string]any, len(values))
	}

	for _, attr := range values {
		k, v := string(attr.Key), attr.Value.AsInterface()
		e.fields[k] = v
	}
}

// Fields returns all the added fields of the event. The returned map is not safe to
// be modified concurrently with calls to AddField or AddAttributes.
func (e *nonSendingEvent) Fields() map[string]any {
	e.fieldsWritingMu.RLock()
	defer e.fieldsWritingMu.RUnlock()

	return e.fields
}

// SetSampleRate overrides the global sample rate for this event.
//
// For nonSendingEvent, this is a no-op.
func (e *nonSendingEvent) SetSampleRate(_ uint) {
	// No-op, since this event is logging-only and does not require sampling
}

// Send dispatches the event to be sent to Honeycomb, sampling if necessary.
//
// For nonSendingEvent, this is a no-op.
func (e *nonSendingEvent) Send() error {
	// No-op, since this event is logging-only and does not actually send data
	return nil
}

var _ Event = &nonSendingEvent{}
