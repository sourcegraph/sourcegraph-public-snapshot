package traceutil

import (
	"net/http"

	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/appdash/httptrace"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func HTTPMiddleware() func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if DefaultCollector == nil {
		return nil
	}
	config := &httptrace.MiddlewareConfig{
		RouteName: func(r *http.Request) string { return httpctx.RouteName(r) },
		SetContextSpan: func(r *http.Request, id appdash.SpanID) {
			SetSpanID(r, id)

			ctx := httpctx.FromRequest(r)
			ctx = NewContext(ctx, id)
			ctx = sourcegraph.WithClientMetadata(ctx, (&Span{SpanID: id}).Metadata())
			httpctx.SetForRequest(r, ctx)
		},
	}
	return httptrace.Middleware(DefaultCollector, config)
}
