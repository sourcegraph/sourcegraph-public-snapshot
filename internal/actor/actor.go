// Package actor provides the structures for representing an actor who has
// access to resources.
package actor

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Actor represents an agent that accesses resources. It can represent an anonymous user, an
// authenticated user, or an internal Sourcegraph service.
//
// Actor can be propagated across services by using actor.HTTPTransport (used by
// httpcli.InternalClientFactory) and actor.HTTPMiddleware. Before assuming this, ensure
// that actor propagation is enabled on both ends of the request.
//
// To learn more about actor propagation, see: https://sourcegraph.com/notebooks/Tm90ZWJvb2s6OTI=
//
// At most one of UID, AnonymousUID, or Internal must be set.
type Actor struct {
	// UID is the unique ID of the authenticated user.
	// Only set if the current actor is an authenticated user.
	UID int32 `json:",omitempty"`

	// AnonymousUID is the user's semi-stable anonymousID from the request cookie
	// or the 'X-Sourcegraph-Actor-Anonymous-UID' request header.
	// Only set if the user is unauthenticated and the request contains an anonymousID.
	AnonymousUID string `json:",omitempty"`

	// Internal is true if the actor represents an internal Sourcegraph service (and is therefore
	// not tied to a specific user).
	Internal bool `json:",omitempty"`

	// SourcegraphOperator indicates whether the actor is a Sourcegraph operator user account.
	SourcegraphOperator bool `json:",omitempty"`

	// FromSessionCookie is whether a session cookie was used to authenticate the actor. It is used
	// to selectively display a logout link (if the actor wasn't authenticated with a session
	// cookie, logout would be ineffective), and for use in telemetry to indicate if a user
	// might have come from a web browser (likely cookie) or directly via the API.
	FromSessionCookie bool `json:"-"`

	// user is populated lazily by (*Actor).User()
	user     *types.User
	userErr  error
	userOnce sync.Once

	// mockUser indicates this user was created in the context of a test.
	mockUser bool
}

// FromUser returns an actor corresponding to the user with the given ID
func FromUser(uid int32) *Actor { return &Actor{UID: uid} }

// FromActualUser returns an actor corresponding to the user with the given ID
func FromActualUser(user *types.User) *Actor {
	a := &Actor{UID: user.ID, user: user, userErr: nil}
	a.userOnce.Do(func() {})
	return a
}

// FromAnonymousUser returns an actor corresponding to an unauthenticated user with the given anonymous ID
func FromAnonymousUser(anonymousUID string) *Actor { return &Actor{AnonymousUID: anonymousUID} }

// FromMockUser returns an actor corresponding to a test user. Do not use outside of tests.
func FromMockUser(uid int32) *Actor { return &Actor{UID: uid, mockUser: true} }

// UIDString is a helper method that returns the UID as a string.
func (a *Actor) UIDString() string { return strconv.Itoa(int(a.UID)) }

func (a *Actor) String() string {
	return fmt.Sprintf("Actor UID %d, internal %t", a.UID, a.Internal)
}

// IsAuthenticated returns true if the Actor is derived from an authenticated user.
func (a *Actor) IsAuthenticated() bool {
	return a != nil && a.UID != 0
}

// IsInternal returns true if the Actor is an internal actor.
func (a *Actor) IsInternal() bool {
	return a != nil && a.Internal
}

// IsMockUser returns true if the Actor is a test user.
func (a *Actor) IsMockUser() bool {
	return a != nil && a.mockUser
}

type userFetcher interface {
	GetByID(context.Context, int32) (*types.User, error)
}

// User returns the expanded types.User for the actor's ID. The ID is expanded to a full
// types.User using the fetcher, which is likely a *database.UserStore.
func (a *Actor) User(ctx context.Context, fetcher userFetcher) (*types.User, error) {
	a.userOnce.Do(func() {
		a.user, a.userErr = fetcher.GetByID(ctx, a.UID)
	})
	if a.user != nil && a.user.ID != a.UID {
		return nil, errors.Errorf("actor UID (%d) and the ID of the cached User (%d) do not match", a.UID, a.user.ID)
	}
	return a.user, a.userErr
}

type contextKey int

const actorKey contextKey = iota

// FromContext returns a new Actor instance from a given context. It always returns a
// non-nil actor.
func FromContext(ctx context.Context) *Actor {
	a, ok := ctx.Value(actorKey).(*Actor)
	if !ok || a == nil {
		return &Actor{}
	}
	return a
}

// WithActor returns a new context with the given Actor instance.
func WithActor(ctx context.Context, a *Actor) context.Context {
	if a != nil && a.UID != 0 {
		trace.User(ctx, a.UID)
	}
	return context.WithValue(ctx, actorKey, a)
}

// WithInternalActor returns a new context with its actor set to be internal.
//
// 🚨 SECURITY: The caller MUST ensure that it performs its own access controls
// or removal of sensitive data.
func WithInternalActor(ctx context.Context) context.Context {
	return context.WithValue(ctx, actorKey, &Actor{Internal: true})
}
