package tracer

import (
	"context"
	"sync/atomic"

	"github.com/sourcegraph/log"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
)

var otelNoOpTracer = oteltrace.NewNoopTracerProvider().Tracer("internal/tracer/no-op")

// shouldTraceTracer only starts a trace if policy.ShouldTrace evaluates to true in
// contexts. It is the equivalent of internal/trace/ot.StartSpanFromContext.
//
// As long as we use both opentracing and OpenTelemetry, we cannot leverage OpenTelemetry
// span processing to implement policy.ShouldTrace, because opentracing does not propagate
// context correctly.
type shouldTraceTracer struct {
	logger log.Logger
	debug  *atomic.Bool

	// tracer is the wrapped tracer implementation.
	tracer oteltrace.Tracer
}

var _ oteltrace.Tracer = &shouldTraceTracer{}

func (t *shouldTraceTracer) Start(ctx context.Context, spanName string, opts ...oteltrace.SpanStartOption) (context.Context, oteltrace.Span) {
	shouldTrace := policy.ShouldTrace(ctx)
	if shouldTrace {
		if t.debug.Load() {
			t.logger.Info("starting span",
				log.Bool("shouldTrace", shouldTrace),
				log.String("spanName", spanName))
		}
		return t.tracer.Start(ctx, spanName, opts...)
	}

	if t.debug.Load() {
		t.logger.Info("starting no-op span",
			log.Bool("shouldTrace", shouldTrace),
			log.String("spanName", spanName))
	}
	return otelNoOpTracer.Start(ctx, spanName, opts...)
}
