package app

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
)

type contextKey int

const (
	requestStartTimeKey contextKey = iota
	githubAuthInfoKey
)

// requestTimerMiddleware times the handling of the HTTP request.
func requestTimerMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	ctx := httpctx.FromRequest(r)

	// Time the handling of this request.
	ctx = context.WithValue(ctx, requestStartTimeKey, time.Now())

	httpctx.SetForRequest(r, ctx)
	next(w, r)
}

func requestStartTimeFromContext(r *http.Request) time.Time {
	tm, _ := httpctx.FromRequest(r).Value(requestStartTimeKey).(time.Time)
	return tm
}
