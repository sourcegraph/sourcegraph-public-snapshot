package trace

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
	nettrace "golang.org/x/net/trace"

	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Trace is a combined version of golang.org/x/net/trace.Trace and
// opentelemetry.Span, applying its various API functions to both
// underlying trace types. Use New to construct one.
type Trace struct {
	family string

	otelSpan oteltrace.Span
	ntTrace  nettrace.Trace
}

// New returns a new Trace with the specified family and title.
func New(ctx context.Context, family, title string, tags ...Tag) (*Trace, context.Context) {
	tr := Tracer{Tracer: ot.GetTracer(ctx)}
	return tr.New(ctx, family, title, tags...)
}

// SetAttributes sets kv as attributes of the Span.
func (t *Trace) SetAttributes(attributes ...attribute.KeyValue) {
	t.otelSpan.SetAttributes(attributes...)
	t.ntTrace.LazyLog(attributesStringer(attributes), false)
}

// AddEvent records an event on this span with the given name and attributes.
//
// Note that it differs from the underlying (oteltrace.Span).AddEvent slightly, and only
// accepts attributes for simplicity, and for ease of adapting to different nettrace.
func (t *Trace) AddEvent(name string, attributes ...attribute.KeyValue) {
	t.otelSpan.AddEvent(name, oteltrace.WithAttributes(attributes...))
	t.ntTrace.LazyLog(attributesStringer(attributes), false)
}

// LazyPrintf evaluates its arguments with fmt.Sprintf each time the
// /debug/requests page is rendered. Any memory referenced by a will be
// pinned until the trace is finished and later discarded.
func (t *Trace) LazyPrintf(format string, a ...any) {
	t.otelSpan.AddEvent("LazyPrintf", oteltrace.WithAttributes(
		attribute.Stringer("message", stringerFunc(func() string {
			return fmt.Sprintf(format, a...)
		})),
	))
	t.ntTrace.LazyPrintf(format, a...)
}

// SetError declares that this trace and span resulted in an error.
func (t *Trace) SetError(err error) {
	if err == nil {
		return
	}

	t.otelSpan.RecordError(err)
	t.otelSpan.SetStatus(codes.Error, err.Error())

	t.ntTrace.LazyPrintf("error: %v", err)
	t.ntTrace.SetError()
}

// SetErrorIfNotContext calls SetError unless err is context.Canceled or
// context.DeadlineExceeded.
func (t *Trace) SetErrorIfNotContext(err error) {
	if errors.IsAny(err, context.Canceled, context.DeadlineExceeded) {
		t.otelSpan.RecordError(err)
		t.ntTrace.LazyPrintf("error: %v", err)
		return
	}

	t.SetError(err)
}

// Finish declares that this trace and span is complete.
// The trace should not be used after calling this method.
func (t *Trace) Finish() {
	t.ntTrace.Finish()
	t.otelSpan.End()
}

/////////////////////
// Deprecated APIs //
/////////////////////

// LogFields logs fields to the opentracing.Span as well as the nettrace.Trace.
//
// Deprecated: Use AddEvent instead.
func (t *Trace) LogFields(fields ...log.Field) {
	t.AddEvent("LogFields", otLogFieldsToOTelAttrs(fields)...)
}

// TagFields adds fields to the opentracing.Span as tags
// as well as as logs to the nettrace.Trace.
//
// Deprecated: Use SetAttributes instead.
func (t *Trace) TagFields(fields ...log.Field) { t.SetAttributes(otLogFieldsToOTelAttrs(fields)...) }
