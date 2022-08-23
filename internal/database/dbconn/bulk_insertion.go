package dbconn

import (
	"context"
)

type key struct{}

var bulkInsertionKey = key{}

// bulkInsertion returns true if the bulkInsertionKey context value is true.
func bulkInsertion(ctx context.Context) bool {
	v, ok := ctx.Value(bulkInsertionKey).(bool)
	if !ok {
		return false
	}
	return v
}

// WithBulkInsertion sets the bulkInsertionKey context value.
func WithBulkInsertion(ctx context.Context, bulkInsertion bool) context.Context {
	return context.WithValue(ctx, bulkInsertionKey, bulkInsertion)
}
