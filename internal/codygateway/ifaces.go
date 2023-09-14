package codygateway

// Actor represents an actor that is making requests to the Cody Gateway.
type Actor interface {
	// GetID returns the unique identifier for this actor.
	GetID() string
	// GetName returns the human-readable name for this actor.
	GetName() string
	// GetSource returns the source of this actor.
	GetSource() ActorSource
}
