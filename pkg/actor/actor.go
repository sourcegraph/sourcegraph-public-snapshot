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

	// Login is the login of the currently authenticated user, if
	// any. It is provided as a convenience and is not guaranteed to
	// be correct (e.g., the user's login can change during the course
	// of a request if the user renames their account). It is also not
	// guaranteed to be populated (many request paths do not populate
	// it, as an optimization to avoid incurring the Users.Get call).
	Login string `json:",omitempty"`

	// Email is the primary email address of the user, if it is known.
	Email string

	// Provider is the ID provider that is the source of truth for this user's identity.
	// It is either the URL of a SSO Provider or "" if the user authenticated via
	// the native authentication flow.
	Provider string

	// AvatarURL is the URL to an avatar image for the user, if it is known.
	AvatarURL string
}

// FromUser returns an actor corresponding to a user
func FromUser(usr *sourcegraph.User) *Actor {
	return &Actor{UID: usr.Auth0ID, Login: usr.Username, Provider: usr.Provider, Email: usr.Email}
}

func (a *Actor) String() string {
	return fmt.Sprintf("Actor UID %s", a.UID)
}

// IsAuthenticated returns true if the Actor is derived from an authenticated user.
func (a *Actor) IsAuthenticated() bool {
	return a.UID != ""
}

func (a *Actor) AuthInfo() *sourcegraph.AuthInfo {
	return &sourcegraph.AuthInfo{
		UID:   a.UID,
		Login: a.Login,
	}
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
	if a != nil && a.Login != "" {
		traceutil.TraceUser(ctx, a.Login)
	}
	return context.WithValue(ctx, actorKey, a)
}
