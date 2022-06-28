package jobutil

import (
	"context"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewSequentialJob will create a job that sequentially runs a list of jobs.
// This is used to implement logic where we might like to order independent
// search operations, favoring results returns by jobs earlier in the list to
// those appearing later in the list. If this job sees a cancellation for a
// child job, it stops executing additional jobs and returns. If ensureUnique is
// true, this job ensures only unique results among all children are sent (if
// two or more jobs send the same result, only the first unique result is sent,
// subsequent similar results are ignored).
func NewSequentialJob(ensureUnique bool, children ...job.Job) job.Job {
	if len(children) == 0 {
		return &NoopJob{}
	}
	if len(children) == 1 {
		return children[0]
	}
	return &SequentialJob{children: children, ensureUnique: ensureUnique}
}

type SequentialJob struct {
	ensureUnique bool
	children     []job.Job
}

func (s *SequentialJob) Name() string {
	return "SequentialJob"
}

func (s *SequentialJob) Tags() []log.Field {
	return []log.Field{
		log.Bool("ensureUnique", s.ensureUnique),
	}
}

func (s *SequentialJob) Run(ctx context.Context, clients job.RuntimeClients, parentStream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, parentStream, finish := job.StartSpan(ctx, parentStream, s)
	defer func() { finish(alert, err) }()

	var maxAlerter search.MaxAlerter
	var errs errors.MultiError

	stream := parentStream
	if s.ensureUnique {
		var mux sync.Mutex
		dedup := result.NewDeduper()

		stream = streaming.StreamFunc(func(event streaming.SearchEvent) {
			mux.Lock()

			results := event.Results[:0]
			for _, match := range event.Results {
				seen := dedup.Seen(match)
				if seen {
					continue
				}
				dedup.Add(match)
				results = append(results, match)
			}
			event.Results = results
			mux.Unlock()
			parentStream.Send(event)
		})
	}

	for _, child := range s.children {
		alert, err := child.Run(ctx, clients, stream)
		if ctx.Err() != nil {
			// Cancellation or Deadline hit implies it's time to stop running jobs.
			return maxAlerter.Alert, errs
		}
		maxAlerter.Add(alert)
		errs = errors.Append(errs, err)
	}
	return maxAlerter.Alert, errs
}

// NewParallelJob will create a job that runs all its child jobs in separate
// goroutines, then waits for all to complete. It returns an aggregated error
// if any of the child jobs failed.
func NewParallelJob(children ...job.Job) job.Job {
	if len(children) == 0 {
		return &NoopJob{}
	}
	if len(children) == 1 {
		return children[0]
	}
	return &ParallelJob{children: children}
}

type ParallelJob struct {
	children []job.Job
}

func (p *ParallelJob) Name() string {
	return "ParallelJob"
}

func (p *ParallelJob) Tags() []log.Field {
	return []log.Field{}
}

func (p *ParallelJob) Run(ctx context.Context, clients job.RuntimeClients, s streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, s, finish := job.StartSpan(ctx, s, p)
	defer func() { finish(alert, err) }()

	var (
		g          errors.Group
		maxAlerter search.MaxAlerter
	)
	for _, child := range p.children {
		child := child
		g.Go(func() error {
			alert, err := child.Run(ctx, clients, s)
			maxAlerter.Add(alert)
			return err
		})
	}
	return maxAlerter.Alert, g.Wait()
}

// NewTimeoutJob creates a new job that is canceled after the
// timeout is hit. The timer starts with `Run()` is called.
func NewTimeoutJob(timeout time.Duration, child job.Job) job.Job {
	if _, ok := child.(*NoopJob); ok {
		return child
	}
	return &TimeoutJob{
		timeout: timeout,
		child:   child,
	}
}

type TimeoutJob struct {
	child   job.Job
	timeout time.Duration
}

func (t *TimeoutJob) Run(ctx context.Context, clients job.RuntimeClients, s streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, s, finish := job.StartSpan(ctx, s, t)
	defer func() { finish(alert, err) }()

	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	return t.child.Run(ctx, clients, s)
}

func (t *TimeoutJob) Name() string {
	return "TimeoutJob"
}

func (t *TimeoutJob) Tags() []log.Field {
	return []log.Field{
		trace.Stringer("timeout", t.timeout),
	}
}

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
	_, ctx, s, finish := job.StartSpan(ctx, s, l)
	defer func() { finish(alert, err) }()

	ctx, s, cancel := streaming.WithLimit(ctx, s, l.limit)
	defer cancel()

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

func (l *LimitJob) Tags() []log.Field {
	return []log.Field{
		log.Int("limit", l.limit),
	}
}

func NewNoopJob() *NoopJob {
	return &NoopJob{}
}

type NoopJob struct{}

func (e *NoopJob) Run(context.Context, job.RuntimeClients, streaming.Sender) (*search.Alert, error) {
	return nil, nil
}

func (e *NoopJob) Name() string { return "NoopJob" }

func (e *NoopJob) Tags() []log.Field {
	return []log.Field{}
}
