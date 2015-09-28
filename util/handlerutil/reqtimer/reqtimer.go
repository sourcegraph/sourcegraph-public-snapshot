// Package reqtimer contains an HTTP middleware that measures the time
// elapsed during the handling of a single request.
package reqtimer

import (
	"net/http"
	"time"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

type contextKey int

const startKey contextKey = iota

// Middleware times the handling of an HTTP request.
func Middleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	ctx := httpctx.FromRequest(r)
	ctx = context.WithValue(ctx, startKey, time.Now())
	httpctx.SetForRequest(r, ctx)
	next(w, r)
}

// Elapsed returns the elapsed duration since request handling began.
func Elapsed(ctx context.Context) time.Duration {
	tm, _ := ctx.Value(startKey).(time.Time)
	return time.Since(tm)
}
