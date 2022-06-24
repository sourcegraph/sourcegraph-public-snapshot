package trace

import (
	"context"
	"strconv"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/log/otfields"
	"github.com/uber/jaeger-client-go"
	nettrace "golang.org/x/net/trace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"

	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// New returns a new Trace with the specified family and title.
func New(ctx context.Context, family, title string, tags ...Tag) (*Trace, context.Context) {
	tr := Tracer{Tracer: ot.GetTracer(ctx)}
	return tr.New(ctx, family, title, tags...)
}

// ID returns a trace ID, if any, found in the given context. If you need both trace and
// span ID, use trace.Context.
func ID(ctx context.Context) string {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return ""
	}
	return IDFromSpan(span)
}

// IDFromSpan returns a trace ID, if any, found in the given span.
func IDFromSpan(span opentracing.Span) string {
	traceCtx := ContextFromSpan(span)
	if traceCtx == nil {
		return ""
	}
	return traceCtx.TraceID
}

// Context retrieves the full trace context, if any, from context - this includes
// both TraceID and SpanID.
func Context(ctx context.Context) *otfields.TraceContext {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return nil
	}
	return ContextFromSpan(span)
}

// Context retrieves the full trace context, if any, from the span - this includes
// both TraceID and SpanID.
func ContextFromSpan(span opentracing.Span) *otfields.TraceContext {
	ddctx, ok := span.Context().(ddtrace.SpanContext)
	if ok {
		return &otfields.TraceContext{
			TraceID: strconv.FormatUint(ddctx.TraceID(), 10),
			SpanID:  strconv.FormatUint(ddctx.SpanID(), 10),
		}
	}

	spanCtx, ok := span.Context().(jaeger.SpanContext)
	if ok {
		return &otfields.TraceContext{
			TraceID: spanCtx.TraceID().String(),
			SpanID:  spanCtx.SpanID().String(),
		}
	}

	return nil
}

type traceContextKey string

const traceKey = traceContextKey("trace")

// contextWithTrace returns a new context.Context that holds a reference to trace's
// SpanContext. External callers should likely use CopyContext, as this properly propagates all
// tracing context from one context to another.
func contextWithTrace(ctx context.Context, tr *Trace) context.Context {
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

// CopyContext copies the tracing-related context items from one context to another and returns that
// context.
func CopyContext(ctx context.Context, from context.Context) context.Context {
	if tr := TraceFromContext(from); tr != nil {
		ctx = contextWithTrace(ctx, tr)
	}
	if shouldTrace := ot.ShouldTrace(from); shouldTrace {
		ctx = ot.WithShouldTrace(ctx, shouldTrace)
	}
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
func (t *Trace) LazyPrintf(format string, a ...any) {
	t.span.LogFields(Printf("log", format, a...))
	t.trace.LazyPrintf(format, a...)
}

// LogFields logs fields to the opentracing.Span
// as well as the nettrace.Trace.
func (t *Trace) LogFields(fields ...log.Field) {
	t.span.LogFields(fields...)
	t.trace.LazyLog(fieldsStringer(fields), false)
}

// TagFields adds fields to the opentracing.Span as tags
// as well as as logs to the nettrace.Trace.
func (t *Trace) TagFields(fields ...log.Field) {
	enc := spanTagEncoder{Span: t.span}
	for _, field := range fields {
		field.Marshal(&enc)
	}
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

// SetErrorIfNotContext calls SetError unless err is context.Canceled or
// context.DeadlineExceeded.
func (t *Trace) SetErrorIfNotContext(err error) {
	if errors.IsAny(err, context.Canceled, context.DeadlineExceeded) {
		t.trace.LazyPrintf("error: %v", err)
		t.span.LogFields(log.Error(err))
		return
	}
	t.SetError(err)
}

// Finish declares that this trace and span is complete.
// The trace should not be used after calling this method.
func (t *Trace) Finish() {
	t.trace.Finish()
	t.span.Finish()
}
