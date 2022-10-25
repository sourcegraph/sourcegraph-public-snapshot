package trace

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
	nettrace "golang.org/x/net/trace"
)

// A Tracer for trace creation, parameterised over an opentelemetry.TracerProvider. Set
// TracerProvider if you don't want to use the global tracer provider, otherwise the
// global TracerProvider is used.
type Tracer struct {
	TracerProvider oteltrace.TracerProvider
}

// New returns a new Trace with the specified family and title.
func (t Tracer) New(ctx context.Context, family, title string, tags ...Tag) (*Trace, context.Context) {
	if t.TracerProvider == nil {
		t.TracerProvider = otel.GetTracerProvider()
	}

	var otelSpan oteltrace.Span
	ctx, otelSpan = t.TracerProvider.
		Tracer("internal/trace").
		Start(ctx, family,
			oteltrace.WithAttributes(attribute.String("title", title)),
			oteltrace.WithAttributes(tagSet(tags).toAttributes()...))

	// Create the nettrace trace to tee to.
	ntTrace := nettrace.New(family, title)

	// Set up the split trace.
	trace := &Trace{
		family:        family,
		oteltraceSpan: otelSpan,
		nettraceTrace: ntTrace,
	}
	if parent := TraceFromContext(ctx); parent != nil {
		ntTrace.LazyPrintf("parent: %s", parent.family)
		trace.family = parent.family + " > " + family
	}
	for _, t := range tags {
		ntTrace.LazyPrintf("%s: %s", t.Key, t.Value)
	}
	return trace, contextWithTrace(ctx, trace)
}

// Tag may be passed when creating a new span. See
// https://github.com/opentracing/specification/blob/master/semantic_conventions.md
// for common tags.
type Tag struct {
	Key   string
	Value string
}

type tagSet []Tag

func (t tagSet) toAttributes() []attribute.KeyValue {
	attributes := make([]attribute.KeyValue, len(t))
	for i, tag := range t {
		attributes[i] = attribute.String(tag.Key, tag.Value)
	}
	return attributes
}
