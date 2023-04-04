package tracer

import (
	"sync/atomic"

	"github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/log"
)

// loggedOTTracer wraps opentracing.Tracer.
type loggedOTTracer struct {
	tracer opentracing.Tracer

	debug  *atomic.Bool
	logger log.Logger
}

var _ opentracing.Tracer = &loggedOTTracer{}

func newLoggedOTTracer(logger log.Logger, tracer opentracing.Tracer, debug *atomic.Bool) *loggedOTTracer {
	return &loggedOTTracer{
		tracer: tracer,
		logger: logger.AddCallerSkip(1),
		debug:  debug,
	}
}

func (t *loggedOTTracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	if t.debug.Load() {
		t.logger.Info("StartSpan",
			log.String("operationName", operationName))
	}
	return t.tracer.StartSpan(operationName, opts...)
}

func (t *loggedOTTracer) Inject(sm opentracing.SpanContext, format any, carrier any) error {
	if t.debug.Load() {
		t.logger.Info("Inject")
	}
	return t.tracer.Inject(sm, format, carrier)
}

func (t *loggedOTTracer) Extract(format any, carrier any) (opentracing.SpanContext, error) {
	if t.debug.Load() {
		t.logger.Info("Extract")
	}
	return t.tracer.Extract(format, carrier)
}
