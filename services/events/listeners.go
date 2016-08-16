package events

import (
	"context"
)

// EventListeners can subscribe to events and perform gRPC operations in
// response to those events. Each listener will receive a context in its
// Start() method which the listener can use to perform gRPC operations
// on the server. The context will be authorized with the scopes requested
// by the listener via its Scopes() method.
type EventListener interface {
	Scopes() []string
	Start(ctx context.Context)
}

// Listeners is a list of event listeners which will be initialized
// at server startup (src serve).
var listeners []EventListener

func RegisterListener(l EventListener) {
	listeners = append(listeners, l)
}

func GetRegisteredListeners() []EventListener {
	return listeners
}
