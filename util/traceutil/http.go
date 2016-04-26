package traceutil

import (
	"net/http"

	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/appdash/httptrace"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
)

func HTTPMiddleware() handlerutil.Middleware {
	if DefaultCollector == nil {
		return nil
	}
	config := &httptrace.MiddlewareConfig{
		RouteName: func(r *http.Request) string {
			// If we have an error, name is an empty string which
			// indicates to httptrace to use a fallback value
			name, _ := httpctx.RouteNameOrError(r)
			return name
		},
		SetContextSpan: func(r *http.Request, id appdash.SpanID) {
			SetSpanID(r, id)

			ctx := httpctx.FromRequest(r)
			ctx = NewContext(ctx, id)
			ctx = sourcegraph.WithClientMetadata(ctx, (&span{spanID: id}).Metadata())
			httpctx.SetForRequest(r, ctx)
		},
	}

	m := httptrace.Middleware(DefaultCollector, config)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m(w, r, next.ServeHTTP)
		})
	}
}
