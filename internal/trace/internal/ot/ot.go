// Package ot (internal/ot) exports internal OpenTelemetry helpers for Sourcegraph's trace
// package.
package ot

import (
	"context"

	"github.com/opentracing/opentracing-go"

	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
)

// StartSpanFromContext starts a span using the tracer returned by GetTracer.
func StartSpanFromContext(ctx context.Context, operationName string, opts ...opentracing.StartSpanOption) (opentracing.Span, context.Context) {
	return StartSpanFromContextWithTracer(ctx, opentracing.GlobalTracer(), operationName, opts...)
}

// StartSpanFromContextWithTracer starts a span using the tracer returned by invoking getTracer with the
// passed-in tracer.
func StartSpanFromContextWithTracer(ctx context.Context, tracer opentracing.Tracer, operationName string, opts ...opentracing.StartSpanOption) (opentracing.Span, context.Context) {
	return opentracing.StartSpanFromContextWithTracer(ctx, GetTracer(ctx, tracer), operationName, opts...)
}

// GetTracer is like GetTracer, but accepts a tracer as an argument. If ShouldTrace returns false,
// it returns the NoopTracer. If it returns true and the passed-in tracer is not nil, it returns the
// passed-in tracer. Otherwise, it returns the global tracer.
func GetTracer(ctx context.Context, tracer opentracing.Tracer) opentracing.Tracer {
	if !policy.ShouldTrace(ctx) {
		return opentracing.NoopTracer{}
	}
	if tracer == nil {
		return opentracing.GlobalTracer()
	}
	return tracer
}
