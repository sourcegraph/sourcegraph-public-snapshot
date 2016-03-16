package pgsql

import (
	"golang.org/x/net/context"
	"gopkg.in/gorp.v1"
	"sourcegraph.com/sourcegraph/sourcegraph/util/traceutil"
)

type contextKey int

const (
	dbhKey contextKey = iota
)

// NewContext creates a new child context with the specified DB
// handle.
func NewContext(ctx context.Context, dbh gorp.SqlExecutor) context.Context {
	return context.WithValue(ctx, dbhKey, dbh)
}

// dbh returns the context's DB handle.
func dbh(ctx context.Context) gorp.SqlExecutor {
	dbh, ok := ctx.Value(dbhKey).(gorp.SqlExecutor)
	if !ok {
		panic("no DB handle set in context")
	}
	return traceutil.SQLExecutor{
		SqlExecutor: dbh,
		Recorder:    traceutil.Recorder(ctx),
	}
}
