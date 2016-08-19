package basictracer

import (
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// Span provides access to the essential details of the span, for use
// by basictracer consumers.  These methods may only be called prior
// to (*opentracing.Span).Finish().
type Span interface {
	opentracing.Span

	// Operation names the work done by this span instance
	Operation() string

	// Start indicates when the span began
	Start() time.Time
}

// Implements the `Span` interface. Created via tracerImpl (see
// `basictracer.New()`).
type spanImpl struct {
	tracer     *tracerImpl
	event      func(SpanEvent)
	sync.Mutex // protects the fields below
	raw        RawSpan
}

var spanPool = &sync.Pool{New: func() interface{} {
	return &spanImpl{}
}}

func (s *spanImpl) reset() {
	s.tracer, s.event = nil, nil
	// Note: Would like to do the following, but then the consumer of RawSpan
	// (the recorder) needs to make sure that they're not holding on to the
	// baggage or logs when they return (i.e. they need to copy if they care):
	//
	//     logs, baggage := s.raw.Logs[:0], s.raw.Baggage
	//     for k := range baggage {
	//         delete(baggage, k)
	//     }
	//     s.raw.Logs, s.raw.Baggage = logs, baggage
	//
	// That's likely too much to ask for. But there is some magic we should
	// be able to do with `runtime.SetFinalizer` to reclaim that memory into
	// a buffer pool when GC considers them unreachable, which should ease
	// some of the load. Hard to say how quickly that would be in practice
	// though.
	s.raw = RawSpan{
		Context: SpanContext{},
	}
}

func (s *spanImpl) SetOperationName(operationName string) opentracing.Span {
	s.Lock()
	defer s.Unlock()
	s.raw.Operation = operationName
	return s
}

func (s *spanImpl) trim() bool {
	return !s.raw.Context.Sampled && s.tracer.options.TrimUnsampledSpans
}

func (s *spanImpl) SetTag(key string, value interface{}) opentracing.Span {
	defer s.onTag(key, value)
	s.Lock()
	defer s.Unlock()
	if key == string(ext.SamplingPriority) {
		if v, ok := value.(uint16); ok {
			s.raw.Context.Sampled = v != 0
			return s
		}
	}
	if s.trim() {
		return s
	}

	if s.raw.Tags == nil {
		s.raw.Tags = opentracing.Tags{}
	}
	s.raw.Tags[key] = value
	return s
}

func (s *spanImpl) LogEvent(event string) {
	s.Log(opentracing.LogData{
		Event: event,
	})
}

func (s *spanImpl) LogEventWithPayload(event string, payload interface{}) {
	s.Log(opentracing.LogData{
		Event:   event,
		Payload: payload,
	})
}

func (s *spanImpl) Log(ld opentracing.LogData) {
	defer s.onLog(ld)
	s.Lock()
	defer s.Unlock()
	if s.trim() || s.tracer.options.DropAllLogs {
		return
	}

	if ld.Timestamp.IsZero() {
		ld.Timestamp = time.Now()
	}

	s.raw.Logs = append(s.raw.Logs, ld)
}

func (s *spanImpl) Finish() {
	s.FinishWithOptions(opentracing.FinishOptions{})
}

func (s *spanImpl) FinishWithOptions(opts opentracing.FinishOptions) {
	finishTime := opts.FinishTime
	if finishTime.IsZero() {
		finishTime = time.Now()
	}
	duration := finishTime.Sub(s.raw.Start)

	s.Lock()
	defer s.Unlock()
	if opts.BulkLogData != nil {
		s.raw.Logs = append(s.raw.Logs, opts.BulkLogData...)
	}
	s.raw.Duration = duration

	s.onFinish(s.raw)
	s.tracer.options.Recorder.RecordSpan(s.raw)

	// Last chance to get options before the span is possbily reset.
	poolEnabled := s.tracer.options.EnableSpanPool
	if s.tracer.options.DebugAssertUseAfterFinish {
		// This makes it much more likely to catch a panic on any subsequent
		// operation since s.tracer is accessed on every call to `Lock`.
		s.reset()
	}

	if poolEnabled {
		spanPool.Put(s)
	}
}

func (s *spanImpl) Tracer() opentracing.Tracer {
	return s.tracer
}

func (s *spanImpl) Context() opentracing.SpanContext {
	return s.raw.Context
}

func (s *spanImpl) SetBaggageItem(key, val string) opentracing.Span {
	s.onBaggage(key, val)
	if s.trim() {
		return s
	}

	s.Lock()
	defer s.Unlock()
	s.raw.Context = s.raw.Context.WithBaggageItem(key, val)
	return s
}

func (s *spanImpl) BaggageItem(key string) string {
	s.Lock()
	defer s.Unlock()
	return s.raw.Context.Baggage[key]
}

func (s *spanImpl) Operation() string {
	return s.raw.Operation
}

func (s *spanImpl) Start() time.Time {
	return s.raw.Start
}
