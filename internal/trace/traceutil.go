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

type traceContextKey string

const traceNameKey = traceContextKey("traceName")

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

// LazyPrintf evaluates its arguments with fmt.Sprintf each time the
// /debug/requests page is rendered. Any memory referenced by a will be
// pinned until the trace is finished and later discarded.
func (t *Trace) LazyPrintf(format string, a ...interface{}) {
	t.LogFields(Printf("log", format, a...))
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

// Printf is an opentracing log.Field which is a LazyLogger. So the format
// string will only be evaluated if the trace is collected. In the case of
// net/trace, it will only be evaluated on page load.
func Printf(key, f string, args ...interface{}) log.Field {
	return log.Lazy(func(fv log.Encoder) {
		fv.EmitString(key, fmt.Sprintf(f, args...))
	})
}

// Stringer is an opentracing log.Field which is a LazyLogger. So the String()
// will only be called if the trace is collected. In the case of net/trace, it
// will only be evaluated on page load.
func Stringer(key string, v fmt.Stringer) log.Field {
	return log.Lazy(func(fv log.Encoder) {
		fv.EmitString(key, v.String())
	})
}
