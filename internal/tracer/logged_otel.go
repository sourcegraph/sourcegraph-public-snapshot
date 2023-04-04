package tracer

import (
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
	return &loggedOtelTracerProvider{logger: logger, provider: provider, debug: debug}
}

// Tracer implements the OpenTelemetry TracerProvider interface. It must do nothing except
// return s.concreteTracer.
func (s *loggedOtelTracerProvider) Tracer(instrumentationName string, opts ...oteltrace.TracerOption) oteltrace.Tracer {
	return s.concreteTracer(instrumentationName, opts...)
}

// concreteTracer generates a concrete shouldTraceTracer OpenTelemetry Tracer implementation, and is used by
// Tracer to implement TracerProvider, and is used by tests to assert against concreteTracer types.
func (s *loggedOtelTracerProvider) concreteTracer(instrumentationName string, opts ...oteltrace.TracerOption) *shouldTraceTracer {
	logger := s.logger
	if s.debug.Load() {
		// Only assign fields to logger in debug mode
		logger = s.logger.With(
			log.String("tracerName", instrumentationName),
			log.String("provider", fmt.Sprintf("%T", s.provider)))
		logger.Info("Tracer")
	}
	return &shouldTraceTracer{
		logger: logger,
		debug:  s.debug,
		tracer: s.provider.Tracer(instrumentationName, opts...),
	}
}
