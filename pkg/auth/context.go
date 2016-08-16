package auth

import "context"

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
