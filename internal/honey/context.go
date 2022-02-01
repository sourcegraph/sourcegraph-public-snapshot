package honey

import (
	"context"
)

type key int

const actorKey key = iota

func WithEvent(ctx context.Context, event Event) context.Context {
	return context.WithValue(ctx, actorKey, event)
}

func FromContext(ctx context.Context) Event {
	if event, ok := ctx.Value(actorKey).(Event); ok {
		return event
	}
	return nil
}
