package tracetest

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	oteltracesdk "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// ConfigureStaticTracerProvider sets up a static tracer provider that generates
// IDs from StaticTraceIDGenerator, and sets up a cleanup on t to restore a
// no-op tracer provider.
//
// This allows code testing tracing integrations to generate predictable, valid
// trace spans.
func ConfigureStaticTracerProvider(t *testing.T) {
	t.Cleanup(func() {
		otel.SetTracerProvider(oteltrace.NewNoopTracerProvider())
	})
	otel.SetTracerProvider(oteltracesdk.NewTracerProvider(
		oteltracesdk.WithIDGenerator(StaticTraceIDGenerator{})))
}

// StaticTraceIDGenerator generates a stable trace and span ID for golden testing.
type StaticTraceIDGenerator struct{}

// NewIDs returns a new trace and span ID.
func (s StaticTraceIDGenerator) NewIDs(ctx context.Context) (oteltrace.TraceID, oteltrace.SpanID) {
	tid, _ := oteltrace.TraceIDFromHex("01020304050607080102040810203040")
	return tid, s.NewSpanID(ctx, tid)
}

// NewSpanID returns a ID for a new span in the trace with traceID.
func (StaticTraceIDGenerator) NewSpanID(ctx context.Context, traceID oteltrace.TraceID) oteltrace.SpanID {
	sid, _ := oteltrace.SpanIDFromHex("0102040810203040")
	return sid
}
