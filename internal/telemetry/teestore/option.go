package teestore

import "context"

type contextKey int

const withoutV1Key contextKey = iota

// WithoutV1 adds a special flag to context that indicates to an underlying
// events teestore.Store that it should not persist the event as a V1 event
// (i.e. event_logs).
//
// This is useful for callsites where the shape of the legacy event must be
// preserved, such that it continues to be logged manually.
func WithoutV1(ctx context.Context) context.Context {
	return context.WithValue(ctx, withoutV1Key, true)
}

func shouldDisableV1(ctx context.Context) bool {
	v, ok := ctx.Value(withoutV1Key).(bool)
	return ok && v
}
