package tracer

import (
	"fmt"
	"sync/atomic"

	"github.com/sourcegraph/log"

	oteltrace "go.opentelemetry.io/otel/trace"
)

// switchableTracer implements otel.TracerProvider, and is used to configure the global
// tracer implementations. It is set as a global tracer so that all opentracing usages
// will end up using this tracer.
//
// The underlying opentracer used is switchable (set via the `set` method), so as to
// support live configuration.
type switchableOtelTracerProvider struct {
	logger       log.Logger
	noopProvider oteltrace.TracerProvider

	v *atomic.Value
}

type otelTracerProvider struct {
	provider oteltrace.TracerProvider
	debug    bool
}

var _ oteltrace.TracerProvider = &switchableOtelTracerProvider{}

func newSwitchableOtelTracerProvider(logger log.Logger) *switchableOtelTracerProvider {
	var v atomic.Value
	v.Store(&otelTracerProvider{
		provider: oteltrace.NewNoopTracerProvider(),
		debug:    false,
	})
	return &switchableOtelTracerProvider{logger: logger, v: &v}
}

func (s *switchableOtelTracerProvider) Tracer(instrumentationName string, opts ...oteltrace.TracerOption) oteltrace.Tracer {
	val := s.v.Load().(*otelTracerProvider) // must be initialized
	if val.debug {
		s.logger.Info("Tracer",
			log.String("provider", fmt.Sprintf("%T", val.provider)))
	}
	return val.provider.Tracer(instrumentationName, opts...)
}

func (s *switchableOtelTracerProvider) set(provider oteltrace.TracerProvider, debug bool) {
	if debug {
		s.logger.Info("set",
			log.String("provider", fmt.Sprintf("%T", provider)))
	}
	if provider == nil {
		provider = oteltrace.NewNoopTracerProvider()
	}
	s.v.Store(&otelTracerProvider{
		provider: provider,
		debug:    debug,
	})
}
