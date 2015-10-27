package events

import (
	"golang.org/x/net/context"
)

type EventListener interface {
	Scopes() []string
	Start(ctx context.Context)
}

// Listeners is a list of event listeners which will be initialized
// at server startup (src serve). Each listener will receive a context
// in its Start() method which the listener can use to perform gRPC operations
// on the server. The context will be authorized with the scopes requested
// by the listener via its Scopes() method.
var Listeners []EventListener
