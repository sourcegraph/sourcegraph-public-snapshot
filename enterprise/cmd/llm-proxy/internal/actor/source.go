package actor

import "context"

// Source is the interface for actor sources.
type Source interface {
	// Get retrieves an actor by an implementation-specific key.
	Get(ctx context.Context, key string) (*Actor, error)
	// Update updates the given actor's state.
	Update(ctx context.Context, actor *Actor)
}
