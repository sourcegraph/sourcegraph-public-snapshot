package tracer

import (
	"context"
	"testing"

	oteltracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
)

func newSampledTracer(t *testing.T) (oteltrace.Tracer, *tracetest.InMemoryExporter) {
	exporter := tracetest.NewInMemoryExporter()
	provider := oteltracesdk.NewTracerProvider(
		oteltracesdk.WithSpanProcessor(oteltracesdk.NewSimpleSpanProcessor(exporter)),
		// use the sampler we want to test
		oteltracesdk.WithSampler(tracePolicySampler{}))
	return provider.Tracer(t.Name()), exporter
}

func TestTracePolicySampler(t *testing.T) {
	// Sanity-check that all trace IDs are unique
	seenTraceIDs := make(map[string]struct{})
	checkTraceID := func(t *testing.T, span oteltrace.Span) {
		tid := span.SpanContext().TraceID().String()
		assert.NotEmpty(t, tid)
		_, exists := seenTraceIDs[tid]
		assert.Falsef(t, exists, "trace ID %q must be unique", tid)
		seenTraceIDs[tid] = struct{}{}
	}

	// Assert that this span is not sampled and not exported.
	assertNotSampled := func(t *testing.T, span oteltrace.Span, exporter *tracetest.InMemoryExporter) {
		checkTraceID(t, span)

		assert.False(t, span.IsRecording(), "span should not be recording")
		assert.False(t, span.SpanContext().IsSampled(), "span should not be sampled")
		assert.True(t, span.SpanContext().IsValid(), "span should be valid")

		// Should not be exported
		span.End()
		assert.Len(t, exporter.GetSpans(), 0)
	}

	// Assert that this span is sampled and gets exported.
	assertSampled := func(t *testing.T, span oteltrace.Span, exporter *tracetest.InMemoryExporter) {
		checkTraceID(t, span)

		assert.True(t, span.IsRecording(), "span should be recording")
		assert.True(t, span.SpanContext().IsSampled(), "span should be sampled")
		assert.True(t, span.SpanContext().IsValid(), "span should be valid")

		// Should be exported
		span.End()
		assert.Len(t, exporter.GetSpans(), 1)
	}

	t.Run("policy.TraceNone", func(t *testing.T) {
		policy.SetTracePolicy(policy.TraceNone) // is default, but explicitly set, just in case

		t.Run("not ShouldTrace", func(t *testing.T) {
			tracer, exporter := newSampledTracer(t)
			_, span := tracer.Start(context.Background(), "Span")
			assertNotSampled(t, span, exporter)
		})

		t.Run("is ShouldTrace", func(t *testing.T) {
			tracer, exporter := newSampledTracer(t)
			ctx := policy.WithShouldTrace(context.Background(), true)
			_, span := tracer.Start(ctx, "Span")
			assertNotSampled(t, span, exporter)
		})
	})

	t.Run("policy.TraceSelective", func(t *testing.T) {
		policy.SetTracePolicy(policy.TraceSelective)
		t.Cleanup(func() { policy.SetTracePolicy(policy.TraceNone) }) // default

		t.Run("not ShouldTrace", func(t *testing.T) {
			tracer, exporter := newSampledTracer(t)
			_, span := tracer.Start(context.Background(), "Span")
			assertNotSampled(t, span, exporter)
		})

		t.Run("is ShouldTrace", func(t *testing.T) {
			tracer, exporter := newSampledTracer(t)
			ctx := policy.WithShouldTrace(context.Background(), true)
			_, span := tracer.Start(ctx, "Span")
			assertSampled(t, span, exporter)
		})
	})

	t.Run("policy.TraceAll", func(t *testing.T) {
		policy.SetTracePolicy(policy.TraceAll)
		t.Cleanup(func() { policy.SetTracePolicy(policy.TraceNone) }) // default

		tracer, exporter := newSampledTracer(t)
		_, span := tracer.Start(context.Background(), "Span")
		assertSampled(t, span, exporter)
	})
}
