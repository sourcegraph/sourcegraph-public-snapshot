package trace

import (
	"context"

	"github.com/sourcegraph/log"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
)

type traceContextKey string

const traceKey = traceContextKey("trace")

// contextWithTrace returns a new context.Context that holds a reference to trace's
// SpanContext. External callers should likely use CopyContext, as this properly propagates all
// tracing context from one context to another.
func contextWithTrace(ctx context.Context, tr *Trace) context.Context {
	ctx = oteltrace.ContextWithSpan(ctx, tr.oteltraceSpan)
	ctx = context.WithValue(ctx, traceKey, tr)
	return ctx
}

// FromContext returns the Trace previously associated with ctx, or
// nil if no such Trace could be found.
func FromContext(ctx context.Context) *Trace {
	tr, _ := ctx.Value(traceKey).(*Trace)
	if tr == nil {
		// There is no Trace in the context, so check for a raw OTel span we can use.
		span := oteltrace.SpanFromContext(ctx)
		if span.IsRecording() {
			tr = &Trace{oteltraceSpan: span}
		}
	}
	return tr
}

// CopyContext copies the tracing-related context items from one context to another and returns that
// context.
func CopyContext(ctx context.Context, from context.Context) context.Context {
	if tr := FromContext(from); tr != nil {
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
	return Context(ctx).TraceID
}

// Context retrieves the full trace context, if any, from context - this includes
// both TraceID and SpanID.
func Context(ctx context.Context) log.TraceContext {
	if otelSpan := oteltrace.SpanContextFromContext(ctx); otelSpan.IsValid() {
		return log.TraceContext{
			TraceID: otelSpan.TraceID().String(),
			SpanID:  otelSpan.SpanID().String(),
		}
	}

	// no span found
	return log.TraceContext{}
}
