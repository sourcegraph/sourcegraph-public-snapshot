package trace

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
	nettrace "golang.org/x/net/trace"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

// EnableNetTrace toggles golang.org/x/net/trace support (exported via
// '/debug/requests' and '/debug/events') in internal/trace through SRC_ENABLE_NET_TRACE.
var EnableNetTrace = env.MustGetBool("SRC_ENABLE_NET_TRACE", true, "Enable golang.org/x/net/trace")

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

	// Create the nettrace trace to tee to. May be left nil if EnableNetTrace
	// is false.
	var ntTrace nettrace.Trace
	if EnableNetTrace {
		ntTrace = nettrace.New(family, title)
	}

	// Set up the split trace.
	trace := &Trace{
		family:        family,
		oteltraceSpan: otelSpan,
		nettraceTrace: ntTrace,
	}
	if parent := TraceFromContext(ctx); parent != nil {
		if ntTrace != nil {
			ntTrace.LazyPrintf("parent: %s", parent.family)
		}
		trace.family = parent.family + " > " + family
	}
	if ntTrace != nil {
		for _, t := range attrs {
			ntTrace.LazyPrintf("%s: %s", t.Key, t.Value)
		}
	}
	return trace, contextWithTrace(ctx, trace)
}
