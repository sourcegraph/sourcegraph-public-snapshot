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

func (s *switchableOtelTracerProvider) Tracer(instrumentationName string, opts ...oteltrace.TracerOption) oteltrace.Tracer {
	val := s.current.Load().(*otelTracerProviderCarrier) // must be initialized
	if val.debug {
		s.logger.Info("Tracer",
			log.String("provider", fmt.Sprintf("%T", val.provider)))
	}
	return val.provider.Tracer(instrumentationName, opts...)
}

func (s *switchableOtelTracerProvider) set(provider oteltrace.TracerProvider, closer io.Closer, debug bool) {
	if debug {
		s.logger.Info("set",
			log.String("provider", fmt.Sprintf("%T", provider)))
	}

	// Shut down previous provider
	if previous := s.current.Load().(*otelTracerProviderCarrier); previous.closer != nil {
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
