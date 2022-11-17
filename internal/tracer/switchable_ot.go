package tracer

import (
	"fmt"
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/log"
)

// switchableOTTracer implements opentracing.Tracer, and is used to configure the global
// tracer implementations. It is set as a global tracer so that all opentracing usages
// will end up using this tracer.
//
// The underlying opentracer used is switchable (set via the `set` method), so as to
// support live configuration.
type switchableOTTracer struct {
	mu     sync.RWMutex
	tracer opentracing.Tracer

	debug  bool
	logger log.Logger
}

var _ opentracing.Tracer = &switchableOTTracer{}

func newSwitchableOTTracer(logger log.Logger) *switchableOTTracer {
	var t opentracing.NoopTracer
	return &switchableOTTracer{
		tracer: t,
		logger: logger.With(log.String("tracer", fmt.Sprintf("%T", t))).AddCallerSkip(1),
	}
}

func (t *switchableOTTracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.debug {
		t.logger.Info("StartSpan",
			log.String("operationName", operationName))
	}
	return t.tracer.StartSpan(operationName, opts...)
}

func (t *switchableOTTracer) Inject(sm opentracing.SpanContext, format any, carrier any) error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.debug {
		t.logger.Info("Inject")
	}
	return t.tracer.Inject(sm, format, carrier)
}

func (t *switchableOTTracer) Extract(format any, carrier any) (opentracing.SpanContext, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.debug {
		t.logger.Info("Extract")
	}
	return t.tracer.Extract(format, carrier)
}

func (t *switchableOTTracer) set(
	logger log.Logger,
	tracer opentracing.Tracer,
	debug bool,
) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.tracer = tracer
	t.debug = debug
	t.logger = logger.With(log.String("tracer", fmt.Sprintf("%T", tracer))).AddCallerSkip(1)

	t.logger.Info("tracer set")
}
