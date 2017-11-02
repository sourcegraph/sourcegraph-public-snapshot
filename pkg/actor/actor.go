package actor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
)

// Actor represents an agent that accesses resources. It can represent
// an anonymous user or a logged-in user.
type Actor struct {
	// UID from the authentication provider. This uniquely identifies
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

	// TODO: add an actor session expiry so if cookie stolen, can't be used for indefinite access. (Not here, somewhere else)

	// GitHubConnected indicates if the actor has a GitHub account connected.
	//
	// DEPRECATED
	GitHubConnected bool

	// GitHubScopes is the list of allowed GitHub API scopes we currently have for the actor.
	//
	// DEPRECATED
	GitHubScopes []string

	// GitHubToken is the token for the GitHub API for this actor.
	// FIXME: It is not nice to store this here, but currently our codebase expects it to be quickly
	// avaialble everywhere.
	//
	// DEPRECATED
	GitHubToken string
}

func (a *Actor) String() string {
	return fmt.Sprintf("Actor UID %s", a.UID)
}

// IsAuthenticated returns true if the Actor is derived from an authenticated user.
func (a *Actor) IsAuthenticated() bool {
	return a.UID != ""
}

func (a *Actor) UserSpec() *sourcegraph.UserSpec {
	return &sourcegraph.UserSpec{
		UID: a.UID,
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

func WithoutActor(ctx context.Context) context.Context {
	return context.WithValue(ctx, actorKey, nil)
}

const HeaderKey = "X-Actor"

// SetTrustedHeader overwrites the entire "X-Actor" header with the actor
// in the context. The actor header may container sensitive information, and as
// such should NEVER be sent to a foreign service. It should not be sent back
// to the client.
func SetTrustedHeader(ctx context.Context, h http.Header) {
	// Remove any existing X-Actor header value that could be provided by an
	// attacker.
	h.Del(HeaderKey)

	// Marshal and store our actor in the header.
	d, err := json.Marshal(FromContext(ctx))
	if err != nil {
		panic(err)
	}
	h.Set(HeaderKey, string(d))
}

// TrustedActorMiddleware is an http.Handler middleware that reads the already
// authenticated actor directly from the "X-Actor" header and sets it in the
// context. No authentication is performed, and thus the HTTP handler should
// NEVER be accessible by a user with control over the "X-Actor" header value.
func TrustedActorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var a Actor
		_ = json.Unmarshal([]byte(r.Header.Get(HeaderKey)), &a)
		next.ServeHTTP(w, r.WithContext(WithActor(r.Context(), &a)))
	})
}
