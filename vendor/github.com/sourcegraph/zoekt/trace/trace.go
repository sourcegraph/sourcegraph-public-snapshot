// Package trace provides a tracing API that in turn invokes both the `golang.org/x/net/trace` API
// and creates an opentracing span if appropriate.
//
// This is similar to the github.com/sourcegraph/sourcegraph/internal/trace package in the main repo,
// and it may make sense to factor both out into a common package at some point.
package trace

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	nettrace "golang.org/x/net/trace"
)

// A Tracer for trace creation, parameterised over an
// opentracing.Tracer. Use this if you don't want to use
// the global tracer.
type Tracer struct {
	Tracer opentracing.Tracer
}

func New(ctx context.Context, family, title string) (*Trace, context.Context) {
	tr := Tracer{Tracer: GetOpenTracer(ctx, nil)}
	return tr.New(ctx, family, title)
}

// New returns a new Trace with the specified family and title.
func (t Tracer) New(ctx context.Context, family, title string) (*Trace, context.Context) {
	span, ctx := StartSpanFromContextWithTracer(
		ctx,
		t.Tracer,
		family,
		opentracing.Tag{Key: "title", Value: title},
	)
	tr := nettrace.New(family, title)
	trace := &Trace{span: span, trace: tr, family: family}
	if parent := TraceFromContext(ctx); parent != nil {
		tr.LazyPrintf("parent: %s", parent.family)
		trace.family = parent.family + " > " + family
	}
	return trace, ContextWithTrace(ctx, trace)
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
	t.span.LogFields(Printf("log", format, a...))
	t.trace.LazyPrintf(format, a...)
}

func (t *Trace) LazyLog(x fmt.Stringer, sensitive bool) {
	t.trace.LazyLog(x, sensitive)
}

// LogFields logs fields to the opentracing.Span
// as well as the nettrace.Trace.
func (t *Trace) LogFields(fields ...log.Field) {
	t.span.LogFields(fields...)
	t.trace.LazyLog(fieldsStringer(fields), false)
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

type traceContextKey string

const traceKey = traceContextKey("trace")

// ContextWithTrace returns a new context.Context that holds a reference to
// trace's SpanContext.
func ContextWithTrace(ctx context.Context, tr *Trace) context.Context {
	ctx = opentracing.ContextWithSpan(ctx, tr.span)
	ctx = context.WithValue(ctx, traceKey, tr)
	return ctx
}

// TraceFromContext returns the Trace previously associated with ctx, or
// nil if no such Trace could be found.
func TraceFromContext(ctx context.Context) *Trace {
	tr, _ := ctx.Value(traceKey).(*Trace)
	return tr
}

// fieldsStringer lazily marshals a slice of log.Field into a string for
// printing in net/trace.
type fieldsStringer []log.Field

func (fs fieldsStringer) String() string {
	var e encoder
	for _, f := range fs {
		f.Marshal(&e)
	}
	return e.Builder.String()
}

// encoder is a log.Encoder used by fieldsStringer.
type encoder struct {
	strings.Builder
	prefixNewline bool
}

func (e *encoder) EmitString(key, value string) {
	if e.prefixNewline {
		// most times encoder is used is for one field
		e.Builder.WriteString("\n")
	}
	if !e.prefixNewline {
		e.prefixNewline = true
	}

	e.Builder.Grow(len(key) + 1 + len(value))
	e.Builder.WriteString(key)
	e.Builder.WriteString(":")
	e.Builder.WriteString(value)
}

func (e *encoder) EmitBool(key string, value bool) {
	e.EmitString(key, strconv.FormatBool(value))
}

func (e *encoder) EmitInt(key string, value int) {
	e.EmitString(key, strconv.Itoa(value))
}

func (e *encoder) EmitInt32(key string, value int32) {
	e.EmitString(key, strconv.FormatInt(int64(value), 10))
}

func (e *encoder) EmitInt64(key string, value int64) {
	e.EmitString(key, strconv.FormatInt(value, 10))
}

func (e *encoder) EmitUint32(key string, value uint32) {
	e.EmitString(key, strconv.FormatUint(uint64(value), 10))
}

func (e *encoder) EmitUint64(key string, value uint64) {
	e.EmitString(key, strconv.FormatUint(value, 10))
}

func (e *encoder) EmitFloat32(key string, value float32) {
	e.EmitString(key, strconv.FormatFloat(float64(value), 'E', -1, 64))
}

func (e *encoder) EmitFloat64(key string, value float64) {
	e.EmitString(key, strconv.FormatFloat(value, 'E', -1, 64))
}

func (e *encoder) EmitObject(key string, value interface{}) {
	e.EmitString(key, fmt.Sprintf("%+v", value))
}

func (e *encoder) EmitLazyLogger(value log.LazyLogger) {
	value(e)
}
