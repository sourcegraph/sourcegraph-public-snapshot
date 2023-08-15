package job

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/atomic"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type finishSpanFunc func(*search.Alert, error)

func StartSpan(ctx context.Context, stream streaming.Sender, job Job) (trace.Trace, context.Context, streaming.Sender, finishSpanFunc) {
	tr, ctx := trace.New(ctx, job.Name())
	tr.SetAttributes(job.Attributes(VerbosityMax)...)

	observingStream := newObservingStream(tr, stream)

	return tr, ctx, observingStream, func(alert *search.Alert, err error) {
		tr.SetError(err)
		if alert != nil {
			tr.SetAttributes(attribute.String("alert", alert.Title))
		}
		tr.SetAttributes(attribute.Int64("total_results", observingStream.totalEvents.Load()))
		tr.End()
	}
}

func newObservingStream(tr trace.Trace, parent streaming.Sender) *observingStream {
	return &observingStream{tr: tr, parent: parent}
}

type observingStream struct {
	tr          trace.Trace
	parent      streaming.Sender
	totalEvents atomic.Int64
}

func (o *observingStream) Send(event streaming.SearchEvent) {
	if l := len(event.Results); l > 0 {
		newTotal := o.totalEvents.Add(int64(l))
		// Only log the first results once. We can rely on reusing the atomic
		// int64 as a "sync.Once" since it is only ever incremented.
		if newTotal == int64(l) {
			o.tr.AddEvent("first results")
		}
	}
	o.parent.Send(event)
}
