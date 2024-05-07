package trace

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// tracerName is the name of the default tracer for the Sourcegraph backend.
const tracerName = "sourcegraph/internal/trace"

// GetTracer returns the default tracer for the Sourcegraph backend.
func GetTracer() oteltrace.Tracer {
	return otel.GetTracerProvider().Tracer(tracerName)
}

// Trace is a light wrapper of opentelemetry.Span. Use New to construct one.
type Trace struct {
	oteltrace.Span // never nil
}

// New returns a new Trace with the specified name in the default tracer.
// For tips on naming, see the OpenTelemetry Span documentation:
// https://opentelemetry.io/docs/specs/otel/trace/api/#span
func New(ctx context.Context, name string, attrs ...attribute.KeyValue) (Trace, context.Context) {
	return NewInTracer(ctx, GetTracer(), name, attrs...)
}

// NewInTracer is the same as New, but uses the given tracer.
func NewInTracer(ctx context.Context, tracer oteltrace.Tracer, name string, attrs ...attribute.KeyValue) (Trace, context.Context) {
	ctx, span := tracer.Start(ctx, name, oteltrace.WithAttributes(attrs...))
	return Trace{span}, ctx
}

// AddEvent records an event on this span with the given name and attributes.
//
// Note that it differs from the underlying (oteltrace.Span).AddEvent slightly, and only
// accepts attributes for simplicity.
func (t Trace) AddEvent(name string, attributes ...attribute.KeyValue) {
	t.Span.AddEvent(name, oteltrace.WithAttributes(attributes...))
}

// SetError declares that this trace and span resulted in an error.
func (t Trace) SetError(err error) {
	if err == nil {
		return
	}

	// Truncate the error string to avoid tracing massive error messages.
	err = truncateError(err, defaultErrorRuneLimit)

	t.RecordError(err)
	t.SetStatus(codes.Error, err.Error())
}

// SetErrorIfNotContext calls SetError unless err is context.Canceled or
// context.DeadlineExceeded.
func (t Trace) SetErrorIfNotContext(err error) {
	if errors.IsAny(err, context.Canceled, context.DeadlineExceeded) {
		err = truncateError(err, defaultErrorRuneLimit)
		t.RecordError(err)
		return
	}

	t.SetError(err)
}

// EndWithErrIfNotContext finishes the span and sets its error value unless it
// is context.Canceled or context.DeadlineExceeded.
//
// It takes a pointer to an error so it can be used directly
// in a defer statement.
func (t Trace) EndWithErrIfNotContext(err *error) {
	t.SetErrorIfNotContext(*err)
	t.End()
}

// EndWithErr finishes the span and sets its error value.
// It takes a pointer to an error so it can be used directly
// in a defer statement.
func (t Trace) EndWithErr(err *error) {
	t.SetError(*err)
	t.End()
}
