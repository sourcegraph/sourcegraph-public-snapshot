package auth

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/env"
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

func DebugMode(ctx context.Context) bool {
	if env.Debug {
		return true
	}
	if ActorFromContext(ctx).Admin {
		return true
	}
	return false
}
