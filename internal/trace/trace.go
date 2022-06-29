package trace

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	nettrace "golang.org/x/net/trace"

	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Trace is a combined version of golang.org/x/net/trace.Trace and
// opentracing.Span. Use New to construct one.
type Trace struct {
	trace  nettrace.Trace
	otSpan opentracing.Span
	family string
}

// New returns a new Trace with the specified family and title.
func New(ctx context.Context, family, title string, tags ...Tag) (*Trace, context.Context) {
	tr := Tracer{Tracer: ot.GetTracer(ctx)}
	return tr.New(ctx, family, title, tags...)
}

// LazyPrintf evaluates its arguments with fmt.Sprintf each time the
// /debug/requests page is rendered. Any memory referenced by a will be
// pinned until the trace is finished and later discarded.
func (t *Trace) LazyPrintf(format string, a ...any) {
	t.otSpan.LogFields(Printf("log", format, a...))
	t.trace.LazyPrintf(format, a...)
}

// LogFields logs fields to the opentracing.Span
// as well as the nettrace.Trace.
func (t *Trace) LogFields(fields ...log.Field) {
	t.otSpan.LogFields(fields...)
	t.trace.LazyLog(fieldsStringer(fields), false)
}

// TagFields adds fields to the opentracing.Span as tags
// as well as as logs to the nettrace.Trace.
func (t *Trace) TagFields(fields ...log.Field) {
	enc := spanTagEncoder{Span: t.otSpan}
	for _, field := range fields {
		field.Marshal(&enc)
	}
	t.trace.LazyLog(fieldsStringer(fields), false)
}

// SetTag lets Trace implement opentrace.Span.SetTag
func (t *Trace) SetTag(key string, value interface{}) {
	t.otSpan.SetTag(key, value)
	t.trace.LazyPrintf("%s: %v", key, value)
}

// SetError declares that this trace and span resulted in an error.
func (t *Trace) SetError(err error) {
	if err == nil {
		return
	}
	t.trace.LazyPrintf("error: %v", err)
	t.trace.SetError()
	t.otSpan.LogFields(log.Error(err))
	ext.Error.Set(t.otSpan, true)
}

// SetErrorIfNotContext calls SetError unless err is context.Canceled or
// context.DeadlineExceeded.
func (t *Trace) SetErrorIfNotContext(err error) {
	if errors.IsAny(err, context.Canceled, context.DeadlineExceeded) {
		t.trace.LazyPrintf("error: %v", err)
		t.otSpan.LogFields(log.Error(err))
		return
	}
	t.SetError(err)
}

// Finish declares that this trace and span is complete.
// The trace should not be used after calling this method.
func (t *Trace) Finish() {
	t.trace.Finish()
	t.otSpan.Finish()
}
