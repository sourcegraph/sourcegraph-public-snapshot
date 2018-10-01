package actor

import (
	"context"
	"fmt"
	"strconv"

	"github.com/sourcegraph/sourcegraph/pkg/trace"
)

// Actor represents an agent that accesses resources. It can represent
// an anonymous user or an authenticated user.
type Actor struct {
	// UID is the unique ID of the authenticated user, or 0 for anonymous actors.
	UID int32 `json:",omitempty"`

	// FromSessionCookie is whether a session cookie was used to authenticate the actor. It is used
	// to selectively display a logout link. (If the actor wasn't authenticated with a session
	// cookie, logout would be ineffective.)
	FromSessionCookie bool `json:"-"`
}

// FromUser returns an actor corresponding to a user
func FromUser(uid int32) *Actor { return &Actor{UID: uid} }

// UIDString is a helper method that returns the UID as a string.
func (a *Actor) UIDString() string { return strconv.Itoa(int(a.UID)) }

func (a *Actor) String() string {
	return fmt.Sprintf("Actor UID %d", a.UID)
}

// IsAuthenticated returns true if the Actor is derived from an authenticated user.
func (a *Actor) IsAuthenticated() bool {
	return a != nil && a.UID != 0
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
	if a != nil && a.UID != 0 {
		trace.TraceUser(ctx, a.UID)
	}
	return context.WithValue(ctx, actorKey, a)
}
