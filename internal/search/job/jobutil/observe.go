package jobutil

import (
	"context"

	"github.com/opentracing/opentracing-go/log"
	"go.uber.org/atomic"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

type observableJob interface {
	Name() string
}

type finishSpanFunc func(*search.Alert, error)

func StartSpan(ctx context.Context, stream streaming.Sender, job observableJob) (*trace.Trace, context.Context, streaming.Sender, finishSpanFunc) {
	tr, ctx := trace.New(ctx, job.Name(), "")

	basicFinish := func(alert *search.Alert, err error) {
		tr.SetError(err)
		if alert != nil {
			tr.TagFields(log.String("alert", alert.Title))
		}
		tr.Finish()
	}

	if ot.ShouldTrace(ctx) {
		// Only wrap the stream if we are actually tracing since the stream
		// wrapper is not zero cost.
		observingStream := newObservingStream(tr, stream)
		return tr, ctx, observingStream, func(alert *search.Alert, err error) {
			tr.LogFields(log.Int64("total_results", observingStream.totalEvents.Load()))
			basicFinish(alert, err)
		}
	}

	return tr, ctx, stream, basicFinish
}

func newObservingStream(tr *trace.Trace, parent streaming.Sender) *observingStream {
	return &observingStream{tr: tr, parent: parent}
}

type observingStream struct {
	tr          *trace.Trace
	parent      streaming.Sender
	totalEvents atomic.Int64
}

func (o *observingStream) Send(event streaming.SearchEvent) {
	if l := len(event.Results); l > 0 {
		o.tr.LogFields(log.Int("results", l))
		o.totalEvents.Add(int64(l))
	}
	o.parent.Send(event)
}
