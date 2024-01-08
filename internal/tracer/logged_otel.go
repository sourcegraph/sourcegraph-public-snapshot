package tracer

import (
	"context"
	"fmt"
	"sync/atomic"

	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/log"
)

// switchableTracer wraps otel.TracerProvider.
type loggedOtelTracerProvider struct {
	logger   log.Logger
	provider oteltrace.TracerProvider
	debug    *atomic.Bool
}

var _ oteltrace.TracerProvider = &loggedOtelTracerProvider{}

func newLoggedOtelTracerProvider(logger log.Logger, provider oteltrace.TracerProvider, debug *atomic.Bool) *loggedOtelTracerProvider {
	return &loggedOtelTracerProvider{logger: logger.AddCallerSkip(1), provider: provider, debug: debug}
}

// Tracer implements the OpenTelemetry TracerProvider interface. It must do nothing except
// return s.concreteTracer.
func (s *loggedOtelTracerProvider) Tracer(instrumentationName string, opts ...oteltrace.TracerOption) oteltrace.Tracer {
	return s.concreteTracer(instrumentationName, opts...)
}

// concreteTracer generates a concrete shouldTraceTracer OpenTelemetry Tracer implementation, and is used by
// Tracer to implement TracerProvider, and is used by tests to assert against concreteTracer types.
func (s *loggedOtelTracerProvider) concreteTracer(instrumentationName string, opts ...oteltrace.TracerOption) *loggedOtelTracer {
	logger := s.logger
	if s.debug.Load() {
		// Only assign fields to logger in debug mode
		logger = s.logger.With(
			log.String("tracerName", instrumentationName),
			log.String("provider", fmt.Sprintf("%T", s.provider)))
		logger.Info("Tracer")
	}
	return &loggedOtelTracer{
		logger: logger,
		debug:  s.debug,
		tracer: s.provider.Tracer(instrumentationName, opts...),
	}
}

type loggedOtelTracer struct {
	logger log.Logger
	debug  *atomic.Bool

	// tracer is the wrapped tracer implementation.
	tracer oteltrace.Tracer
}

var _ oteltrace.Tracer = &loggedOtelTracer{}

func (t *loggedOtelTracer) Start(ctx context.Context, spanName string, opts ...oteltrace.SpanStartOption) (context.Context, oteltrace.Span) {
	ctx, span := t.tracer.Start(ctx, spanName, opts...)
	if t.debug.Load() {
		t.logger.Info("Start",
			log.String("spanName", spanName),
			log.Bool("isRecording", span.IsRecording()),
			log.Bool("isSampled", span.SpanContext().IsSampled()),
			log.Bool("isValid", span.SpanContext().IsValid()))
	}
	return ctx, span
}
