package auth

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/idkey"
)

type key int

const (
	actorKey key = iota
	ticketsKey
	repoCheckerKey
	repoCheckerStartedKey
	oauth2ConfigKey
)

func ActorFromContext(ctx context.Context) Actor {
	a, _ := ctx.Value(actorKey).(Actor)
	return a
}

// LoginFromContext returns the login of the currently authenticated
// user (in the context). If there is no such user, or if the login
// can't be determined without incurring an additional DB lookup, ("",
// false) is returned.
//
// Because the user's login is not always stored (for performance
// reasons), this func shouldn't be used to check if there is an
// authenticated user; call ActorFromContext(ctx).IsAuthenticated()
// instead.
func LoginFromContext(ctx context.Context) (string, bool) {
	a := ActorFromContext(ctx)
	if !a.IsAuthenticated() || a.Login == "" {
		return "", false
	}
	return a.Login, true
}

func WithActor(ctx context.Context, a Actor) context.Context {
	return context.WithValue(ctx, actorKey, a)
}

func TicketsFromContext(ctx context.Context) []Perm {
	tix, _ := ctx.Value(ticketsKey).([]Perm)
	return tix
}

func WithTickets(ctx context.Context, tix []Perm) context.Context {
	return context.WithValue(ctx, ticketsKey, tix)
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
