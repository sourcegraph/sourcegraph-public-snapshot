package dbconn

import (
	"context"
)

type key struct{}

var bulkInsertionKey = key{}

// isBulkInsertion returns true if the bulkInsertionKey context value is true.
func isBulkInsertion(ctx context.Context) bool {
	v, ok := ctx.Value(bulkInsertionKey).(bool)
	if !ok {
		return false
	}
	return v
}

// WithBulkInsertion sets whether or not a bulk insertion is occurring within this context.
func WithBulkInsertion(ctx context.Context, bulkInsertion bool) context.Context {
	return context.WithValue(ctx, bulkInsertionKey, bulkInsertion)
}
