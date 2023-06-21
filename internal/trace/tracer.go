package trace

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// A Tracer for trace creation, parameterised over an opentelemetry.TracerProvider. Set
// TracerProvider if you don't want to use the global tracer provider, otherwise the
// global TracerProvider is used.
type Tracer struct {
	TracerProvider oteltrace.TracerProvider
}

// New returns a new Trace with the specified family and title. Must be closed with Finish().
func (t Tracer) New(ctx context.Context, family, title string, attrs ...attribute.KeyValue) (*Trace, context.Context) {
	if t.TracerProvider == nil {
		t.TracerProvider = otel.GetTracerProvider()
	}

	var otelSpan oteltrace.Span
	ctx, otelSpan = t.TracerProvider.
		Tracer("sourcegraph/internal/trace").
		Start(ctx, family,
			oteltrace.WithAttributes(attribute.String("title", title)),
			oteltrace.WithAttributes(attrs...))

	// Set up the split trace.
	trace := &Trace{
		family:        family,
		oteltraceSpan: otelSpan,
	}
	if parent := TraceFromContext(ctx); parent != nil {
		trace.family = parent.family + " > " + family
	}
	return trace, contextWithTrace(ctx, trace)
}
