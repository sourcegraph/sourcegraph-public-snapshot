package tracer

import (
	"fmt"
	"io"
	"sync/atomic"

	"github.com/sourcegraph/log"

	oteltrace "go.opentelemetry.io/otel/trace"
)

// switchableTracer implements otel.TracerProvider, and is used to configure the global
// tracer implementations. It is set as a global tracer so that all opentracing usages
// will end up using this tracer.
//
// The underlying tracer provider used is switchable (set via the `set` method), so as to
// support live configuration.
type switchableOtelTracerProvider struct {
	logger log.Logger

	// current caries the *otelTracerProvider currently associated with this provider.
	current *atomic.Value
}

type otelTracerProviderCarrier struct {
	provider oteltrace.TracerProvider
	closer   io.Closer
	debug    bool
}

var _ oteltrace.TracerProvider = &switchableOtelTracerProvider{}

func newSwitchableOtelTracerProvider(logger log.Logger) *switchableOtelTracerProvider {
	var v atomic.Value
	v.Store(&otelTracerProviderCarrier{
		provider: oteltrace.NewNoopTracerProvider(),
		debug:    false,
	})
	return &switchableOtelTracerProvider{logger: logger, current: &v}
}

// Tracer implements the OpenTelemetry TracerProvider interface. It must do nothing except
// return s.concreteTracer.
func (s *switchableOtelTracerProvider) Tracer(instrumentationName string, opts ...oteltrace.TracerOption) oteltrace.Tracer {
	return s.concreteTracer(instrumentationName, opts...)
}

// concreteTracer generates a concrete concreteTracer OpenTelemetry Tracer implementation, and is used by
// Tracer to implement TracerProvider, and is used by tests to assert against concreteTracer types.
func (s *switchableOtelTracerProvider) concreteTracer(instrumentationName string, opts ...oteltrace.TracerOption) *shouldTraceTracer {
	carrier := s.loadCarrier()

	logger := s.logger
	if carrier.debug {
		// Only assign fields to logger in debug mode
		logger = s.logger.With(
			log.String("tracerName", instrumentationName),
			log.String("provider", fmt.Sprintf("%T", carrier.provider)))
		logger.Info("Tracer")
	}
	return &shouldTraceTracer{
		logger: logger,
		debug:  carrier.debug,
		tracer: carrier.provider.Tracer(instrumentationName, opts...),
	}
}

// loadCarrier retrieves the carrier struct that configures the current TracerProvider and
// pipeline closer. The current value must already be initialized in the constructor.
func (s *switchableOtelTracerProvider) loadCarrier() *otelTracerProviderCarrier {
	return s.current.Load().(*otelTracerProviderCarrier)
}

func (s *switchableOtelTracerProvider) set(provider oteltrace.TracerProvider, closer io.Closer, debug bool) {
	if debug {
		s.logger.Info("set",
			log.String("provider", fmt.Sprintf("%T", provider)))
	}

	// Shut down previous provider
	if previous := s.loadCarrier(); previous.closer != nil {
		go previous.closer.Close() // non-blocking
	}

	// Ensure we default to a valid tracer
	if provider == nil {
		provider = oteltrace.NewNoopTracerProvider()
	}

	// Update the value
	s.current.Store(&otelTracerProviderCarrier{
		provider: provider,
		closer:   closer,
		debug:    debug,
	})
}
