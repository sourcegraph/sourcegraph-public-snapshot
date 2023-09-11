package telemetry

// eventName defines the name of an event for telemetry purposes.
//
// This is a private type, requiring the values to be declared in-package
// and preventing strings from being cast to this type.
type eventName string

// All event names in Sourcegraph's Go services.
const (
	EventExample = "Example"
)
