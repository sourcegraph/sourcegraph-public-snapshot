package traceutil

import (
	"net/http"

	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/appdash/httptrace"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httputil/httpctx"
)

func HTTPMiddleware(next http.Handler) http.Handler {
	if DefaultCollector == nil {
		return next
	}
	config := &httptrace.MiddlewareConfig{
		RouteName: func(r *http.Request) string {
			// If we have an error, name is an empty string which
			// indicates to httptrace to use a fallback value
			name, _ := httpctx.RouteNameOrError(r)
			return name
		},
		SetContextSpan: func(r *http.Request, id appdash.SpanID) *http.Request {
			ctx := r.Context()
			ctx = NewContext(ctx, id)
			ctx = sourcegraph.WithClientMetadata(ctx, (&span{spanID: id}).Metadata())
			return r.WithContext(ctx)
		},
	}

	m := httptrace.Middleware(DefaultCollector, config)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m(w, r, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Appdash-Trace", SpanIDFromContext(r.Context()).Trace.String())
			next.ServeHTTP(w, r)
		})
	})
}
