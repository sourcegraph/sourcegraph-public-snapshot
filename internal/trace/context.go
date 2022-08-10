package trace

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/log"
	"github.com/uber/jaeger-client-go"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
)

type traceContextKey string

const traceKey = traceContextKey("trace")

// contextWithTrace returns a new context.Context that holds a reference to trace's
// SpanContext. External callers should likely use CopyContext, as this properly propagates all
// tracing context from one context to another.
func contextWithTrace(ctx context.Context, tr *Trace) context.Context {
	ctx = opentracing.ContextWithSpan(ctx, tr.span)
	ctx = context.WithValue(ctx, traceKey, tr)
	return ctx
}

// TraceFromContext returns the Trace previously associated with ctx, or
// nil if no such Trace could be found.
func TraceFromContext(ctx context.Context) *Trace {
	tr, _ := ctx.Value(traceKey).(*Trace)
	return tr
}

// CopyContext copies the tracing-related context items from one context to another and returns that
// context.
func CopyContext(ctx context.Context, from context.Context) context.Context {
	if tr := TraceFromContext(from); tr != nil {
		ctx = contextWithTrace(ctx, tr)
	}
	if shouldTrace := policy.ShouldTrace(from); shouldTrace {
		ctx = policy.WithShouldTrace(ctx, shouldTrace)
	}
	return ctx
}

// ID returns a trace ID, if any, found in the given context. If you need both trace and
// span ID, use trace.Context.
func ID(ctx context.Context) string {
	span := Context(ctx)
	if span == nil {
		return ""
	}
	return span.TraceID
}

// Context retrieves the full trace context, if any, from context - this includes
// both TraceID and SpanID.
func Context(ctx context.Context) *log.TraceContext {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return nil
	}

	// try Jaeger ("opentracing") span
	if jaegerSpan, ok := span.Context().(jaeger.SpanContext); ok {
		return &log.TraceContext{
			TraceID: jaegerSpan.TraceID().String(),
			SpanID:  jaegerSpan.SpanID().String(),
		}
	}

	// try bridged OpenTelemetry span
	if otelSpan := oteltrace.SpanFromContext(ctx).SpanContext(); otelSpan.IsValid() {
		return &log.TraceContext{
			TraceID: otelSpan.TraceID().String(),
			SpanID:  otelSpan.SpanID().String(),
		}
	}

	// no span found
	return nil
}
