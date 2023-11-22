package jobutil

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/atomic"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewLimitJob creates a new job that is canceled after the result limit
// is hit. Whenever an event is sent down the stream, the result count
// is incremented by the number of results in that event, and if it reaches
// the limit, the context is canceled.
func NewLimitJob(limit int, child job.Job) job.Job {
	if _, ok := child.(*NoopJob); ok {
		return child
	}
	return &LimitJob{
		limit: limit,
		child: child,
	}
}

type LimitJob struct {
	child job.Job
	limit int
}

func (l *LimitJob) Run(ctx context.Context, clients job.RuntimeClients, s streaming.Sender) (alert *search.Alert, err error) {
	tr, ctx, s, finish := job.StartSpan(ctx, s, l)
	defer func() { finish(alert, err) }()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	s = newLimitStream(l.limit, s, func() {
		tr.AddEvent("limit hit, canceling child context")
		cancel()
	})

	alert, err = l.child.Run(ctx, clients, s)
	if errors.Is(err, context.Canceled) {
		// Ignore context canceled errors
		err = nil
	}
	return alert, err

}

func (l *LimitJob) Name() string {
	return "LimitJob"
}

func (l *LimitJob) Attributes(v job.Verbosity) (res []attribute.KeyValue) {
	switch v {
	case job.VerbosityMax:
		fallthrough
	case job.VerbosityBasic:
		res = append(res,
			attribute.Int("limit", l.limit),
		)
	}
	return res
}

func (l *LimitJob) Children() []job.Describer {
	return []job.Describer{l.child}
}

func (l *LimitJob) MapChildren(fn job.MapFunc) job.Job {
	cp := *l
	cp.child = job.Map(l.child, fn)
	return &cp
}

type limitStream struct {
	s          streaming.Sender
	onLimitHit context.CancelFunc
	remaining  atomic.Int64
}

func (s *limitStream) Send(event streaming.SearchEvent) {
	count := int64(event.Results.ResultCount())

	// Avoid limit checks if no change to result count.
	if count == 0 {
		s.s.Send(event)
		return
	}

	// Get the remaining count before and after sending this event
	after := s.remaining.Sub(count)
	before := after + count

	// Check if the event needs truncating before being sent
	if after < 0 {
		limit := before
		if before < 0 {
			limit = 0
		}
		event.Results.Limit(int(limit))
	}

	// Send the maybe-truncated event. We want to always send the event
	// even if we truncate it to zero results in case it has stats on it
	// that we care about it.
	s.s.Send(event)

	// Send the IsLimitHit event and call cancel exactly once. This will
	// only trigger when the result count of an event causes us to cross
	// the zero-remaining threshold.
	if after <= 0 && before > 0 {
		s.s.Send(streaming.SearchEvent{Stats: streaming.Stats{IsLimitHit: true}})
		s.onLimitHit()
	}
}

// newLimitStream returns a child Stream of parent. The child stream passes on all events
// to the parent until the limit has been hit. When the limit is hit, it will send a limit
// hit event on the parent stream and call the onLimitHit callback, which can be used
// to, e.g., cancel a context.
func newLimitStream(limit int, parent streaming.Sender, onLimitHit func()) streaming.Sender {
	stream := &limitStream{onLimitHit: onLimitHit, s: parent}
	stream.remaining.Store(int64(limit))
	return stream
}
