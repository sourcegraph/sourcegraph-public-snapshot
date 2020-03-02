package trace

import (
	"context"
	"fmt"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	nettrace "golang.org/x/net/trace"
)

// SpanURL returns the URL to the tracing UI for the given span. The span must be non-nil.
var SpanURL = func(span opentracing.Span) string {
	return "#tracer-not-enabled"
}

// New returns a new Trace with the specified family and title.
func New(ctx context.Context, family, title string) (*Trace, context.Context) {
	tr := Tracer{Tracer: opentracing.GlobalTracer()}
	return tr.New(ctx, family, title)
}

// A Tracer for trace creation, parameterised over an
// opentracing.Tracer. Use this if you don't want to use
// the global tracer.
type Tracer struct {
	Tracer opentracing.Tracer
}

// New returns a new Trace with the specified family and title.
func (t Tracer) New(ctx context.Context, family, title string) (*Trace, context.Context) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(
		ctx,
		t.Tracer,
		family,
		opentracing.Tag{Key: "title", Value: title},
	)
	family, ctx = nameWithParents(ctx, family)
	tr := nettrace.New(family, title)
	return &Trace{span: span, trace: tr, family: family}, ctx
}

const traceNameKey = "traceName"

func nameWithParents(ctx context.Context, name string) (string, context.Context) {
	prefix, _ := ctx.Value(traceNameKey).(string)
	name = prefix + name
	return name, context.WithValue(ctx, traceNameKey, name+" > ")
}

// ContextWithTrace returns a new context.Context that holds a reference to
// trace's SpanContext.
func ContextWithTrace(ctx context.Context, tr *Trace) context.Context {
	ctx = opentracing.ContextWithSpan(ctx, tr.span)
	ctx = context.WithValue(ctx, traceNameKey, tr.family)
	return ctx
}

// Trace is a combined version of golang.org/x/net/trace.Trace and
// opentracing.Span. Use New to construct one.
type Trace struct {
	trace  nettrace.Trace
	span   opentracing.Span
	family string
}

// LazyLog adds x to the net/trace event log. It will be evaluated each time
// the /debug/requests page is rendered. Any memory referenced by x will be
// pinned until the trace is finished and later discarded.
//
// NOTE: It will not log to the opentracing.Span
func (t *Trace) LazyLog(x fmt.Stringer, sensitive bool) {
	t.trace.LazyLog(x, sensitive)
}

// LazyPrintf evaluates its arguments with fmt.Sprintf each time the
// /debug/requests page is rendered. Any memory referenced by a will be
// pinned until the trace is finished and later discarded.
//
// NOTE: It will not log to the opentracing.Span
func (t *Trace) LazyPrintf(format string, a ...interface{}) {
	t.trace.LazyPrintf(format, a...)
}

// LogFields logs fields to the opentracing.Span
// as well as the nettrace.Trace.
func (t *Trace) LogFields(fields ...log.Field) {
	t.span.LogFields(fields...)
	for _, f := range fields {
		t.trace.LazyLog(f, false)
	}
}

// SetError declares that this trace and span resulted in an error.
func (t *Trace) SetError(err error) {
	if err == nil {
		return
	}
	t.trace.LazyPrintf("error: %v", err)
	t.trace.SetError()
	t.span.LogFields(log.Error(err))
	ext.Error.Set(t.span, true)
}

// Finish declares that this trace and span is complete.
// The trace should not be used after calling this method.
func (t *Trace) Finish() {
	t.trace.Finish()
	t.span.Finish()
}
