// Package ot wraps github.com/opentracing/opentracing-go and
// github.com./opentracing-contrib/go-stdlib with selective tracing behavior that is toggled on and
// off with the presence of a context item (uses context.Context). This context item is propagated
// across API boundaries through a HTTP header (X-Sourcegraph-Should-Trace).
package ot

import (
	"context"
	"net/http"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"

	internalot "github.com/sourcegraph/sourcegraph/internal/trace/internal/ot"
	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
)

// GetTracer returns the tracer to use for the given context. If ShouldTrace returns true, it
// returns the global tracer. Otherwise, it returns the NoopTracer.
//
// TODO OpenTelemetry considerations: https://github.com/sourcegraph/sourcegraph/issues/27386
func GetTracer(ctx context.Context) opentracing.Tracer {
	return internalot.GetTracer(ctx, opentracing.GlobalTracer())
}

// HTTPMiddleware wraps the handler with the following:
//
// - If the HTTP header, X-Sourcegraph-Should-Trace, is set to a truthy value, set the
//   shouldTraceKey context.Context value to true
// - github.com/opentracing-contrib/go-stdlib/nethttp.HTTPMiddleware, which creates a new span to track
//   the request handler from the global tracer.
//
// TODO OpenTelemetry considerations: https://github.com/sourcegraph/sourcegraph/issues/27386
func HTTPMiddleware(h http.Handler, opts ...nethttp.MWOption) http.Handler {
	return MiddlewareWithTracer(opentracing.GlobalTracer(), h)
}

// MiddlewareWithTracer is like Middleware, but uses the specified tracer instead of the global
// tracer.
//
// TODO OpenTelemetry considerations: https://github.com/sourcegraph/sourcegraph/issues/27386
func MiddlewareWithTracer(tr opentracing.Tracer, h http.Handler, opts ...nethttp.MWOption) http.Handler {
	nethttpMiddleware := nethttp.Middleware(tr, h, append([]nethttp.MWOption{
		nethttp.MWSpanFilter(func(r *http.Request) bool {
			return policy.ShouldTrace(r.Context())
		}),
	}, opts...)...)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var trace bool
		switch policy.GetTracePolicy() {
		case policy.TraceSelective:
			trace = policy.RequestWantsTracing(r)
		case policy.TraceAll:
			trace = true
		default:
			trace = false
		}
		nethttpMiddleware.ServeHTTP(w, r.WithContext(policy.WithShouldTrace(r.Context(), trace)))
	})
}
