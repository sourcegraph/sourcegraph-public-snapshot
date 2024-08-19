package web

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/zoekt"
	"github.com/sourcegraph/zoekt/query"
	"github.com/sourcegraph/zoekt/trace"
)

func NewTraceAwareSearcher(s zoekt.Streamer) zoekt.Streamer {
	return traceAwareSearcher{Searcher: s}
}

// traceAwareSearcher wraps a zoekt.Searcher instance so that the tracing context item is set in the
// context. This context item toggles on trace collection via the
// github.com/sourcegraph/zoekt/trace/ot package.
type traceAwareSearcher struct {
	Searcher zoekt.Streamer
}

func (s traceAwareSearcher) Search(
	ctx context.Context,
	q query.Q,
	opts *zoekt.SearchOptions,
) (*zoekt.SearchResult, error) {
	ctx = trace.WithOpenTracingEnabled(ctx, opts.Trace)
	spanContext := trace.SpanContextFromContext(ctx)
	if opts.Trace && spanContext != nil {
		var span opentracing.Span
		span, ctx = opentracing.StartSpanFromContext(ctx, "zoekt.traceAwareSearcher.Search", opentracing.ChildOf(spanContext))
		defer span.Finish()
	}
	return s.Searcher.Search(ctx, q, opts)
}

func (s traceAwareSearcher) StreamSearch(
	ctx context.Context,
	q query.Q,
	opts *zoekt.SearchOptions,
	sender zoekt.Sender,
) error {
	ctx = trace.WithOpenTracingEnabled(ctx, opts.Trace)
	spanContext := trace.SpanContextFromContext(ctx)
	if opts.Trace && spanContext != nil {
		var span opentracing.Span
		span, ctx = opentracing.StartSpanFromContext(ctx, "zoekt.traceAwareSearcher.StreamSearch", opentracing.ChildOf(spanContext))
		defer span.Finish()
	}
	return s.Searcher.StreamSearch(ctx, q, opts, sender)
}

func (s traceAwareSearcher) List(ctx context.Context, q query.Q, opts *zoekt.ListOptions) (*zoekt.RepoList, error) {
	return s.Searcher.List(ctx, q, opts)
}
func (s traceAwareSearcher) Close()         { s.Searcher.Close() }
func (s traceAwareSearcher) String() string { return s.Searcher.String() }
