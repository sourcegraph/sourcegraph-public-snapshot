package trace

import (
	"context"

	"github.com/opentracing/opentracing-go"
)

type key int

const (
	enableOpenTracingKey key = iota
)

// isOpenTracingEnabled returns true if the enableOpenTracingKey context value is true.
func isOpenTracingEnabled(ctx context.Context) bool {
	v, ok := ctx.Value(enableOpenTracingKey).(bool)
	if !ok {
		return false
	}
	return v
}

func WithOpenTracingEnabled(ctx context.Context, enableOpenTracing bool) context.Context {
	return context.WithValue(ctx, enableOpenTracingKey, enableOpenTracing)
}

// GetOpenTracer returns the tracer to actually use depending on whether isOpenTracingEnabled(ctx)
// returns true or false. If false, this returns the NoopTracer.
func GetOpenTracer(ctx context.Context, tracer opentracing.Tracer) opentracing.Tracer {
	if !isOpenTracingEnabled(ctx) {
		return opentracing.NoopTracer{}
	}
	if tracer == nil {
		return opentracing.GlobalTracer()
	}
	return tracer
}

// StartSpanFromContext starts a span using the tracer returned by invoking GetOpenTracer with the
// passed-in tracer.
func StartSpanFromContextWithTracer(ctx context.Context, tracer opentracing.Tracer, operationName string, opts ...opentracing.StartSpanOption) (opentracing.Span, context.Context) {
	return opentracing.StartSpanFromContextWithTracer(ctx, GetOpenTracer(ctx, tracer), operationName, opts...)
}

// StartSpanFromContext starts a span using the tracer returned by GetOpenTracer.
func StartSpanFromContext(ctx context.Context, operationName string, opts ...opentracing.StartSpanOption) (opentracing.Span, context.Context) {
	return StartSpanFromContextWithTracer(ctx, GetOpenTracer(ctx, nil), operationName, opts...)
}
