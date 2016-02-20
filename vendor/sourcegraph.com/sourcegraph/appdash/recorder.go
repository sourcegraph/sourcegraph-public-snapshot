package appdash

import (
	"fmt"
	"sync"
)

// A Recorder is associated with a span and records annotations on the
// span by sending them to a collector.
type Recorder struct {
	SpanID // the span ID that annotations are about

	collector Collector // the collector to send to

	errors   []error    // errors since the last call to Errors
	errorsMu sync.Mutex // protects errors
}

// NewRecorder creates a new recorder for the given span and
// collector. If c is nil, NewRecorder panics.
func NewRecorder(span SpanID, c Collector) *Recorder {
	if c == nil {
		panic("Collector is nil")
	}
	return &Recorder{
		SpanID:    span,
		collector: c,
	}
}

// Child creates a new Recorder with the same collector and a new
// child SpanID whose parent is this recorder's SpanID.
func (r *Recorder) Child() *Recorder {
	return NewRecorder(NewSpanID(r.SpanID), r.collector)
}

// Name sets the name of this span.
func (r *Recorder) Name(name string) {
	r.Event(spanName{name})
}

// Msg records a Msg event (an event with a human-readable message) on
// the span.
func (r *Recorder) Msg(msg string) {
	r.Event(Msg(msg))
}

// Log records a Log event (an event with the current timestamp and a
// human-readable message) on the span.
func (r *Recorder) Log(msg string) {
	r.Event(Log(msg))
}

// Event records any event that implements the Event, TimespanEvent, or
// TimestampedEvent interfaces.
func (r *Recorder) Event(e Event) {
	as, err := MarshalEvent(e)
	if err != nil {
		r.error("Event", err)
		return
	}
	r.Annotation(as...)
}

// Annotation records raw annotations on the span.
func (r *Recorder) Annotation(as ...Annotation) {
	if err := r.failsafeAnnotation(as...); err != nil {
		r.error("Annotation", err)
	}
}

// Annotation records raw annotations on the span.
func (r *Recorder) failsafeAnnotation(as ...Annotation) error {
	return r.collector.Collect(r.SpanID, as...)
}

// Errors returns all errors encountered by the Recorder since the
// last call to Errors. After calling Errors, the Recorder's list of
// errors is emptied.
func (r *Recorder) Errors() []error {
	r.errorsMu.Lock()
	errs := r.errors
	r.errors = nil
	r.errorsMu.Unlock()
	return errs
}

func (r *Recorder) error(method string, err error) {
	as, _ := MarshalEvent(Log(fmt.Sprintf("Recorder.%s error: %s", method, err)))
	r.failsafeAnnotation(as...)
	r.errorsMu.Lock()
	r.errors = append(r.errors, err)
	r.errorsMu.Unlock()
}
