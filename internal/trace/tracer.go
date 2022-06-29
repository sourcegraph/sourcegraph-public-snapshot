package trace

import (
	"context"

	"github.com/opentracing/opentracing-go"
	nettrace "golang.org/x/net/trace"

	internalot "github.com/sourcegraph/sourcegraph/internal/trace/internal/ot"
)

// A Tracer for trace creation, parameterised over an
// opentracing.Tracer. Use this if you don't want to use
// the global tracer.
type Tracer struct {
	Tracer opentracing.Tracer
}

// New returns a new Trace with the specified family and title.
func (t Tracer) New(ctx context.Context, family, title string, tags ...Tag) (*Trace, context.Context) {
	span, ctx := internalot.StartSpanFromContextWithTracer(
		ctx,
		t.Tracer,
		family,
		tagsOpt{title: title, tags: tags},
	)
	tr := nettrace.New(family, title)
	trace := &Trace{otSpan: span, trace: tr, family: family}
	if parent := TraceFromContext(ctx); parent != nil {
		tr.LazyPrintf("parent: %s", parent.family)
		trace.family = parent.family + " > " + family
	}
	for _, t := range tags {
		tr.LazyPrintf("%s: %s", t.Key, t.Value)
	}
	return trace, contextWithTrace(ctx, trace)
}

// Tag may be passed when creating a new span. See
// https://github.com/opentracing/specification/blob/master/semantic_conventions.md
// for common tags.
type Tag struct {
	Key   string
	Value string
}

// tagsOpt is an opentracing.StartSpanOption which applies all the tags
type tagsOpt struct {
	tags  []Tag
	title string
}

// Apply satisfies the StartSpanOption interface.
func (t tagsOpt) Apply(o *opentracing.StartSpanOptions) {
	if len(t.tags) == 0 && t.title == "" {
		return
	}
	if o.Tags == nil {
		o.Tags = make(map[string]any, len(t.tags)+1)
	}
	if t.title != "" {
		o.Tags["title"] = t.title
	}
	for _, t := range t.tags {
		o.Tags[t.Key] = t.Value
	}
}
