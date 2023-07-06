package trace

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Trace is a wrapper of opentelemetry.Span. Use New to construct one.
type Trace struct {
	// oteltraceSpan is always set.
	oteltraceSpan oteltrace.Span
}

// New returns a new Trace with the specified name.
// For tips on naming, see the OpenTelemetry Span documentation:
// https://opentelemetry.io/docs/specs/otel/trace/api/#span
func New(ctx context.Context, name string, attrs ...attribute.KeyValue) (Trace, context.Context) {
	tracer := Tracer{TracerProvider: otel.GetTracerProvider()}
	tr, ctx := tracer.New(ctx, name, attrs...)
	return tr, ctx
}

// SetAttributes sets kv as attributes of the Span.
func (t *Trace) SetAttributes(attributes ...attribute.KeyValue) {
	t.oteltraceSpan.SetAttributes(attributes...)
}

// AddEvent records an event on this span with the given name and attributes.
//
// Note that it differs from the underlying (oteltrace.Span).AddEvent slightly, and only
// accepts attributes for simplicity.
func (t *Trace) AddEvent(name string, attributes ...attribute.KeyValue) {
	t.oteltraceSpan.AddEvent(name, oteltrace.WithAttributes(attributes...))
}

// SetError declares that this trace and span resulted in an error.
func (t *Trace) SetError(err error) {
	if err == nil {
		return
	}

	// Truncate the error string to avoid tracing massive error messages.
	err = truncateError(err, defaultErrorRuneLimit)

	t.oteltraceSpan.RecordError(err)
	t.oteltraceSpan.SetStatus(codes.Error, err.Error())
}

// SetErrorIfNotContext calls SetError unless err is context.Canceled or
// context.DeadlineExceeded.
func (t *Trace) SetErrorIfNotContext(err error) {
	if errors.IsAny(err, context.Canceled, context.DeadlineExceeded) {
		err = truncateError(err, defaultErrorRuneLimit)
		t.oteltraceSpan.RecordError(err)
		return
	}

	t.SetError(err)
}

// End declares that this trace and span is complete.
// The trace should not be used after calling this method.
func (t *Trace) End() {
	t.oteltraceSpan.End()
}

// EndWithErr finishes the span and sets its error value.
// It takes a pointer to an error so it can be used directly
// in a defer statement.
func (t *Trace) EndWithErr(err *error) {
	t.SetError(*err)
	t.End()
}
