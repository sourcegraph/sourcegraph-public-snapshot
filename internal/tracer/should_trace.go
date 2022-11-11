package tracer

import (
	"context"

	"github.com/sourcegraph/log"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
)

var otelNoOpTracer = oteltrace.NewNoopTracerProvider().Tracer("internal/tracer/no-op")

// shouldTraceTracer only starts a trace if policy.ShouldTrace evaluates to true in
// contexts.
type shouldTraceTracer struct {
	logger log.Logger
	debug  bool

	// tracer is the wrapped tracer implementation.
	tracer oteltrace.Tracer
}

var _ oteltrace.Tracer = &shouldTraceTracer{}

func (t *shouldTraceTracer) Start(ctx context.Context, spanName string, opts ...oteltrace.SpanStartOption) (context.Context, oteltrace.Span) {
	shouldTrace := policy.ShouldTrace(ctx)
	if shouldTrace {
		if t.debug {
			t.logger.Info("starting span",
				log.Bool("shouldTrace", shouldTrace),
				log.String("spanName", spanName))
		}
		return t.tracer.Start(ctx, spanName, opts...)
	}
	if t.debug {
		t.logger.Info("starting no-op span",
			log.Bool("shouldTrace", shouldTrace),
			log.String("spanName", spanName))
	}
	return otelNoOpTracer.Start(ctx, spanName, opts...)
}
