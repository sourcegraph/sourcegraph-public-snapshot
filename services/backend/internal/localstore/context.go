package localstore

import (
	"context"

	"gopkg.in/gorp.v1"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
)

type contextKey int

const (
	appDBHKey contextKey = iota
	graphDBHKey
)

// WithAppDBH creates a new child context with the specified app DB
// handle.
func WithAppDBH(ctx context.Context, appDBH gorp.SqlExecutor) context.Context {
	return context.WithValue(ctx, appDBHKey, appDBH)
}

// WithGraphDBH creates a new child context with the specified graph DB
// handle.
func WithGraphDBH(ctx context.Context, graphDBH gorp.SqlExecutor) context.Context {
	return context.WithValue(ctx, graphDBHKey, graphDBH)
}

// appDBH returns the context's app DB handle.
func appDBH(ctx context.Context) gorp.SqlExecutor {
	dbh, ok := ctx.Value(appDBHKey).(gorp.SqlExecutor)
	if !ok {
		panic("no app DB handle set in context")
	}
	return traceutil.SQLExecutor{
		SqlExecutor: dbh,
		Recorder:    traceutil.Recorder(ctx),
	}
}

// graphDBH returns the context's app DB handle.
func graphDBH(ctx context.Context) gorp.SqlExecutor {
	dbh, ok := ctx.Value(graphDBHKey).(gorp.SqlExecutor)
	if !ok {
		panic("no graph DB handle set in context")
	}
	return traceutil.SQLExecutor{
		SqlExecutor: dbh,
		Recorder:    traceutil.Recorder(ctx),
	}
}
