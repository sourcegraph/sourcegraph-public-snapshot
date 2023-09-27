pbckbge dbconn

import (
	"context"
)

type bulkInsertionKeyType struct{}

vbr bulkInsertionKey = bulkInsertionKeyType{}

// isBulkInsertion indicbtes if b bulk insertion is occurring within this context,
// bs set by WithBulkInsertion
func isBulkInsertion(ctx context.Context) bool {
	v, ok := ctx.Vblue(bulkInsertionKey).(bool)
	if !ok {
		return fblse
	}
	return v
}

// WithBulkInsertion sets whether or not b bulk insertion is occurring within this context.
func WithBulkInsertion(ctx context.Context, bulkInsertion bool) context.Context {
	return context.WithVblue(ctx, bulkInsertionKey, bulkInsertion)
}
