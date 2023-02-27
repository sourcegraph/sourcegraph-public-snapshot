// Package ot wraps github.com/opentracing/opentracing-go and
// github.com./opentracing-contrib/go-stdlib with selective tracing behavior that is toggled on and
// off with the presence of a context item (uses context.Context). This context item is propagated
// across API boundaries through a HTTP header (X-Sourcegraph-Should-Trace).
package ot

import (
	"context"

	"github.com/opentracing/opentracing-go"

	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
)

// Deprecated: Use otel.Tracer(...) from go.opentelemetry.io/otel instead.
//
// GetTracer returns the tracer to use for the given context. If ShouldTrace returns true, it
// returns the global tracer. Otherwise, it returns the NoopTracer.
func GetTracer(ctx context.Context) opentracing.Tracer {
	return getTracer(ctx, opentracing.GlobalTracer())
}

// Deprecated: Use otel.Tracer(...).Start() from go.opentelemetry.io/otel or trace.New(...)
// from github.com/sourcegraph/sourcegraph/internal/trace instead.
//
// StartSpanFromContext starts a span using the tracer returned by GetTracer.
func StartSpanFromContext(ctx context.Context, operationName string, opts ...opentracing.StartSpanOption) (opentracing.Span, context.Context) {
	return StartSpanFromContextWithTracer(ctx, opentracing.GlobalTracer(), operationName, opts...)
}

// Deprecated: Use otel.Tracer(...).Start() from go.opentelemetry.io/otel or trace.New(...)
// from github.com/sourcegraph/sourcegraph/internal/trace instead.
//
// StartSpanFromContextWithTracer starts a span using the tracer returned by invoking getTracer with the
// passed-in tracer.
func StartSpanFromContextWithTracer(ctx context.Context, tracer opentracing.Tracer, operationName string, opts ...opentracing.StartSpanOption) (opentracing.Span, context.Context) {
	return opentracing.StartSpanFromContextWithTracer(ctx, getTracer(ctx, tracer), operationName, opts...)
}

// getTracer is like GetTracer, but accepts a tracer as an argument. If ShouldTrace returns false,
// it returns the NoopTracer. If it returns true and the passed-in tracer is not nil, it returns the
// passed-in tracer. Otherwise, it returns the global tracer.
func getTracer(ctx context.Context, tracer opentracing.Tracer) opentracing.Tracer {
	if !policy.ShouldTrace(ctx) {
		return opentracing.NoopTracer{}
	}
	if tracer == nil {
		return opentracing.GlobalTracer()
	}
	return tracer
}
