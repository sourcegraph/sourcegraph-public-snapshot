package auth

import (
	"context"
	"fmt"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

// Actor represents an agent that accesses resources. It can represent
// an anonymous user or a logged-in user.
type Actor struct {
	UID string `json:",omitempty"`

	// Login is the login of the currently authenticated user, if
	// any. It is provided as a convenience and is not guaranteed to
	// be correct (e.g., the user's login can change during the course
	// of a request if the user renames their account). It is also not
	// guaranteed to be populated (many request paths do not populate
	// it, as an optimization to avoid incurring the Users.Get call).
	Login string `json:",omitempty"`

	// Scope is a set of authorized scopes that the actor has
	// access to on the given server.
	Scope map[string]bool `json:",omitempty"`

	// Email is the primary email address of the user.
	Email string

	// AvatarURL is the URL to an avatar image for the user.
	AvatarURL string

	// GitHubConnected indicates if the actor has a GitHub account connected.
	GitHubConnected bool

	// GitHubScopes is the list of allowed GitHub API scopes we currently have for the actor.
	GitHubScopes []string

	// GitHubToken is the token for the GitHub API for this actor.
	// FIXME: It is not nice to store this here, but currently our codebase expects it to be quickly
	// avaialble everywhere.
	GitHubToken string

	// GoogleConnected indicates if the actor has a Google account connected.
	GoogleConnected bool

	// GoogleScopes is the list of allowed Google API scopes we currently have for the actor.
	GoogleScopes []string
}

func (a *Actor) String() string {
	return fmt.Sprintf("Actor UID %s (scope=%v)", a.UID, a.Scope)
}

// IsAuthenticated returns true if the Actor is derived from an authenticated user.
func (a *Actor) IsAuthenticated() bool {
	return a.UID != ""
}

// HasScope returns a boolean indicating whether this actor has the
// given scope.
func (a *Actor) HasScope(s string) bool {
	hasScope, ok := a.Scope[s]
	return ok && hasScope
}

func (a *Actor) UserSpec() *sourcegraph.UserSpec {
	return &sourcegraph.UserSpec{
		UID: a.UID,
	}
}

func (a *Actor) User() *sourcegraph.User {
	if a.UID == "" {
		return nil
	}
	return &sourcegraph.User{
		UID:       a.UID,
		Login:     a.Login,
		AvatarURL: a.AvatarURL,
	}
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

func ActorFromContext(ctx context.Context) *Actor {
	a, ok := ctx.Value(actorKey).(*Actor)
	if !ok {
		return &Actor{}
	}
	return a
}

func WithActor(ctx context.Context, a *Actor) context.Context {
	return context.WithValue(ctx, actorKey, a)
}

func WithoutActor(ctx context.Context) context.Context {
	return context.WithValue(ctx, actorKey, nil)
}
