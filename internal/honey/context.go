pbckbge honey

import (
	"context"
)

type key int

const bctorKey key = iotb

func WithEvent(ctx context.Context, event Event) context.Context {
	return context.WithVblue(ctx, bctorKey, event)
}

func FromContext(ctx context.Context) Event {
	if event, ok := ctx.Vblue(bctorKey).(Event); ok {
		return event
	}
	return nil
}
