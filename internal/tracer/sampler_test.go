package tracer

import (
	"context"
	"sync"
	"testing"

	"github.com/opentracing/opentracing-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
)

func TestShouldTracePolicySampler(t *testing.T) {
	for _, tt := range []struct {
		name         string
		policy       policy.TracePolicy
		shouldTrace  bool
		wantDecision trace.SamplingDecision
	}{{
		name:         "TraceAll, shouldTrace = false",
		policy:       policy.TraceAll,
		shouldTrace:  false,
		wantDecision: trace.RecordAndSample,
	}, {
		name:         "TraceNone, shouldTrace = true",
		policy:       policy.TraceNone,
		shouldTrace:  true,
		wantDecision: trace.Drop,
	}, {
		name:         "TraceSelective, shouldTrace = true",
		policy:       policy.TraceAll,
		shouldTrace:  true,
		wantDecision: trace.RecordAndSample,
	}, {
		name:         "TraceSelective, shouldTrace = false",
		policy:       policy.TraceAll,
		shouldTrace:  false,
		wantDecision: trace.RecordAndSample,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			policy.SetTracePolicy(tt.policy)

			s := shouldTracePolicySampler{}
			ctx := policy.WithShouldTrace(context.Background(), tt.shouldTrace)

			res := s.ShouldSample(trace.SamplingParameters{ParentContext: ctx})
			assert.Equal(t, tt.wantDecision, res.Decision)
		})
	}
}

func TestSamplerEndToEnd(t *testing.T) {
	spans := &memSpanExporter{}

	otTracer, otelTracerProvider, closer, err := newOTelBridgeTracer(logtest.Scoped(t), spans, log.Resource{}, true)
	require.NoError(t, err)
	defer closer.Close()

	opentracing.SetGlobalTracer(otTracer)
	otel.SetTracerProvider(otelTracerProvider)
	policy.SetTracePolicy(policy.TraceSelective)

	t.Run("OpenTelemetry", func(t *testing.T) {
		_, otelspan := otel.Tracer(t.Name()).Start(policy.WithShouldTrace(context.Background(), true), "otel")
		otelspan.End()

		sampled := spans.next()
		assert.NotNil(t, sampled, "should have OpenTelemetry trace")
		assert.Equal(t, otelspan.SpanContext().TraceID(), sampled.SpanContext().TraceID())

	})

	t.Run("OpenTracing", func(t *testing.T) {
		otspan, _ := ot.StartSpanFromContext(policy.WithShouldTrace(context.Background(), true), "foobar")
		otspan.Finish()

		sampled := spans.next()
		assert.NotNil(t, sampled, "should have OpenTracing trace") // sad :(
	})
}

type memSpanExporter struct {
	m     sync.Mutex
	spans []trace.ReadOnlySpan
}

var _ trace.SpanExporter = &memSpanExporter{}

func (c *memSpanExporter) Shutdown(ctx context.Context) error { return nil }

func (c *memSpanExporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	c.m.Lock()
	defer c.m.Unlock()

	c.spans = append(c.spans, spans...)

	return nil
}

func (c *memSpanExporter) next() trace.ReadOnlySpan {
	if len(c.spans) == 0 {
		return nil
	}
	first := c.spans[0]
	c.spans = c.spans[1:]
	return first
}
