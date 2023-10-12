package trace

import (
	"context"

	"github.com/sourcegraph/log"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
)

// FromContext returns the Trace previously associated with ctx.
func FromContext(ctx context.Context) Trace {
	return Trace{oteltrace.SpanFromContext(ctx)}
}

// CopyContext copies the tracing-related context items from one context to another and returns that
// context.
func CopyContext(ctx context.Context, from context.Context) context.Context {
	ctx = oteltrace.ContextWithSpan(ctx, oteltrace.SpanFromContext(from))
	ctx = policy.WithShouldTrace(ctx, policy.ShouldTrace(from))
	return ctx
}

// BackgroundContext returns a background context with the same tracing
// policy and trace ID associated with from, if available. It is a convenience
// alias for CopyContext(context.Background(), from).
//
// Using context.Background() is desirable in scenarios where we don't want an
// action to be cancelled if the parent action is cancelled, while retaining a
// trace hierarchy.
//
// 🚨 SECURITY: The returned context does NOT retain any other context baggage,
// including Actor context.
//
// 🔔 TODO(go1.21): Replace all callsites with https://pkg.go.dev/context#WithoutCancel
// when we upgrade to Go .21
func BackgroundContext(from context.Context) context.Context {
	return CopyContext(context.Background(), from)
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
