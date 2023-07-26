package trace

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Trace is a combined version of opentelemetry.Span and (optionally)
// golang.org/x/net/trace.Trace, applying its various API functions to both
// underlying trace types. Use New to construct one.
type Trace struct {
	// oteltraceSpan is always set.
	oteltraceSpan oteltrace.Span
}

// New returns a new Trace with the specified name.
// For tips on naming, see the OpenTelemetry Span documentation:
// https://opentelemetry.io/docs/specs/otel/trace/api/#span
func New(ctx context.Context, name string, attrs ...attribute.KeyValue) (*Trace, context.Context) {
	tr := Tracer{TracerProvider: otel.GetTracerProvider()}
	return tr.New(ctx, name, attrs...)
}

// SetAttributes sets kv as attributes of the Span.
func (t *Trace) SetAttributes(attributes ...attribute.KeyValue) {
	t.oteltraceSpan.SetAttributes(attributes...)
}

// AddEvent records an event on this span with the given name and attributes.
//
// Note that it differs from the underlying (oteltrace.Span).AddEvent slightly, and only
// accepts attributes for simplicity, and for ease of adapting to nettrace.
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

// Finish declares that this trace and span is complete.
// The trace should not be used after calling this method.
func (t *Trace) Finish() {
	t.oteltraceSpan.End()
}

// FinishWithErr finishes the span and sets its error value.
// It takes a pointer to an error so it can be used directly
// in a defer statement.
func (t *Trace) FinishWithErr(err *error) {
	t.SetError(*err)
	t.Finish()
}
