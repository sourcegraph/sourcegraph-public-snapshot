package pgsql

import (
	"github.com/sqs/modl"
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/util/traceutil"
)

type contextKey int

const (
	dbhKey contextKey = iota
)

// NewContext creates a new child context with the specified DB
// handle.
func NewContext(ctx context.Context, dbh modl.SqlExecutor) context.Context {
	return context.WithValue(ctx, dbhKey, dbh)
}

// dbh returns the context's DB handle.
func dbh(ctx context.Context) modl.SqlExecutor {
	dbh, ok := ctx.Value(dbhKey).(modl.SqlExecutor)
	if !ok {
		panic("no DB handle set in context")
	}
	return traceutil.SQLExecutor{
		SqlExecutor: dbh,
		Recorder:    traceutil.Recorder(ctx),
	}
}
