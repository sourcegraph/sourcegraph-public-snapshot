package tracer

import (
	"fmt"
	"io"
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/log"
)

// switchableTracer implements opentracing.Tracer, and is used to configure the global
// tracer implementations. It is set as a global tracer so that all opentracing usages
// will end up using this tracer.
//
// The underlying opentracer used is switchable (set via the `set` method), so as to
// support live configuration.
type switchableTracer struct {
	mu           sync.RWMutex
	tracer       opentracing.Tracer
	tracerCloser io.Closer

	log    bool
	logger log.Logger
}

var _ opentracing.Tracer = &switchableTracer{}

func newSwitchableOTTracer(logger log.Logger) *switchableTracer {
	var t opentracing.NoopTracer
	return &switchableTracer{
		tracer: t,
		logger: logger.With(log.String("tracer", fmt.Sprintf("%T", t))).AddCallerSkip(1),
	}
}

func (t *switchableTracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.log {
		t.logger.Info("StartSpan",
			log.String("operationName", operationName))
	}
	return t.tracer.StartSpan(operationName, opts...)
}

func (t *switchableTracer) Inject(sm opentracing.SpanContext, format any, carrier any) error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.log {
		t.logger.Info("Inject")
	}
	return t.tracer.Inject(sm, format, carrier)
}

func (t *switchableTracer) Extract(format any, carrier any) (opentracing.SpanContext, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.log {
		t.logger.Info("Extract")
	}
	return t.tracer.Extract(format, carrier)
}

func (t *switchableTracer) set(
	logger log.Logger,
	tracer opentracing.Tracer,
	tracerCloser io.Closer,
	shouldLog bool,
) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if tc := t.tracerCloser; tc != nil {
		// Close the old tracerCloser outside the critical zone
		go tc.Close()
	}

	t.tracerCloser = tracerCloser
	t.tracer = tracer
	t.log = shouldLog
	t.logger = logger.With(log.String("tracer", fmt.Sprintf("%T", tracer))).AddCallerSkip(1)

	t.logger.Info("tracer set")
}
