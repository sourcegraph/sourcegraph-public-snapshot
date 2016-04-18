package auth

import (
	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/auth/idkey"
)

type key int

const (
	actorKey key = iota
)

func ActorFromContext(ctx context.Context) Actor {
	a, _ := ctx.Value(actorKey).(Actor)
	return a
}

func WithActor(ctx context.Context, a Actor) context.Context {
	return context.WithValue(ctx, actorKey, a)
}

// IsAuthenticated returns whether the context has an authenticated
// user. If the context has only an authenticated registered API
// client (but no user), IsAuthenticated returns false.
func IsAuthenticated(ctx context.Context) bool {
	return ActorFromContext(ctx).IsAuthenticated()
}

// ClientID retrieves the server's OAuth2 client ID from the
// context. It assumes that the server's ID key was previously stored
// in the context (using idkey.NewContext).
func ClientID(ctx context.Context) string {
	return idkey.FromContext(ctx).ID
}
