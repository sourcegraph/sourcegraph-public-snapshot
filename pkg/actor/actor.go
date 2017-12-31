package actor

import (
	"context"
	"fmt"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
)

// Actor represents an agent that accesses resources. It can represent
// an anonymous user or a logged-in user.
type Actor struct {
	// UID is the ID from the authentication provider. This uniquely identifies
	// the actor's user within the context of the actor's provider.
	UID string `json:",omitempty"`

	// Provider is the ID provider that is the source of truth for this user's identity.
	// It is either the URL of a SSO Provider or "" if the user authenticated via
	// the native authentication flow.
	Provider string
}

// FromUser returns an actor corresponding to a user
func FromUser(usr *sourcegraph.User) *Actor {
	return &Actor{UID: usr.AuthID, Provider: usr.Provider}
}

func (a *Actor) String() string {
	return fmt.Sprintf("Actor UID %s", a.UID)
}

// IsAuthenticated returns true if the Actor is derived from an authenticated user.
func (a *Actor) IsAuthenticated() bool {
	return a.UID != ""
}

type key int

const (
	actorKey key = iota
)

func FromContext(ctx context.Context) *Actor {
	a, ok := ctx.Value(actorKey).(*Actor)
	if !ok || a == nil {
		return &Actor{}
	}
	return a
}

func WithActor(ctx context.Context, a *Actor) context.Context {
	if a != nil && a.UID != "" {
		traceutil.TraceUser(ctx, a.UID)
	}
	return context.WithValue(ctx, actorKey, a)
}
