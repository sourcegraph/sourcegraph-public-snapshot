package job

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewPriorityJob creates a combinator job from a required job and an
// optional job. When run, PriorityJob runs the required job and the
// optional job in parallel, waits for the required job to complete, then gives
// the optional job a short additional amount of time (currently 100ms) before
// canceling the optional job.
func NewPriorityJob(required Job, optional Job) Job {
	if _, ok := optional.(*noopJob); ok {
		return required
	}
	return &PriorityJob{
		required: required,
		optional: optional,
	}
}

type PriorityJob struct {
	required Job
	optional Job
}

func (r *PriorityJob) Name() string {
	return "PriorityJob"
}

func (r *PriorityJob) Run(ctx context.Context, db database.DB, s streaming.Sender) (alert *search.Alert, err error) {
	tr, ctx := jobutil.StartSpan(ctx, r)
	defer func() { jobutil.FinishSpan(tr, alert, err) }()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	start := time.Now()

	var (
		maxAlerter    search.MaxAlerter
		optionalGroup errors.Group
		requiredGroup errors.Group
	)
	requiredGroup.Go(func() error {
		alert, err := r.required.Run(ctx, db, s)
		maxAlerter.Add(alert)
		return err
	})
	optionalGroup.Go(func() error {
		alert, err := r.optional.Run(ctx, db, s)
		maxAlerter.Add(alert)
		return err
	})

	var errs error
	if err := requiredGroup.Wait(); err != nil {
		errs = errors.Append(errs, err)
	}
	tr.LazyPrintf("required group completed")

	// Give optional searches some minimum budget in case required searches return quickly.
	// Cancel all remaining searches after this minimum budget.
	budget := 100 * time.Millisecond
	elapsed := time.Since(start)
	time.AfterFunc(budget-elapsed, cancel)

	if err := optionalGroup.Wait(); err != nil {
		errs = errors.Append(errs, err)
	}
	tr.LazyPrintf("optional group completed")

	return maxAlerter.Alert, errs
}

// NewParallelJob will create a job that runs all its child jobs in separate
// goroutines, then waits for all to complete. It returns an aggregated error
// if any of the child jobs failed.
func NewParallelJob(children ...Job) Job {
	if len(children) == 0 {
		return &noopJob{}
	}
	if len(children) == 1 {
		return children[0]
	}
	return &ParallelJob{children: children}
}

type ParallelJob struct {
	children []Job
}

func (p *ParallelJob) Name() string {
	return "ParallelJob"
}

func (p *ParallelJob) Run(ctx context.Context, db database.DB, s streaming.Sender) (alert *search.Alert, err error) {
	tr, ctx := jobutil.StartSpan(ctx, p)
	defer func() { jobutil.FinishSpan(tr, alert, err) }()

	var (
		g          errors.Group
		maxAlerter search.MaxAlerter
	)
	for _, child := range p.children {
		child := child
		g.Go(func() error {
			alert, err := child.Run(ctx, db, s)
			maxAlerter.Add(alert)
			return err
		})
	}
	return maxAlerter.Alert, g.Wait()
}

// NewTimeoutJob creates a new job that is canceled after the
// timeout is hit. The timer starts with `Run()` is called.
func NewTimeoutJob(timeout time.Duration, child Job) Job {
	if _, ok := child.(*noopJob); ok {
		return child
	}
	return &TimeoutJob{
		timeout: timeout,
		child:   child,
	}
}

type TimeoutJob struct {
	child   Job
	timeout time.Duration
}

func (t *TimeoutJob) Run(ctx context.Context, db database.DB, s streaming.Sender) (alert *search.Alert, err error) {
	tr, ctx := jobutil.StartSpan(ctx, t)
	defer func() { jobutil.FinishSpan(tr, alert, err) }()

	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	return t.child.Run(ctx, db, s)
}

func (t *TimeoutJob) Name() string {
	return "TimeoutJob"
}

// NewLimitJob creates a new job that is canceled after the result limit
// is hit. Whenever an event is sent down the stream, the result count
// is incremented by the number of results in that event, and if it reaches
// the limit, the context is canceled.
func NewLimitJob(limit int, child Job) Job {
	if _, ok := child.(*noopJob); ok {
		return child
	}
	return &LimitJob{
		limit: limit,
		child: child,
	}
}

type LimitJob struct {
	child Job
	limit int
}

func (l *LimitJob) Run(ctx context.Context, db database.DB, s streaming.Sender) (alert *search.Alert, err error) {
	tr, ctx := jobutil.StartSpan(ctx, l)
	defer func() { jobutil.FinishSpan(tr, alert, err) }()

	ctx, s, cancel := streaming.WithLimit(ctx, s, l.limit)
	defer cancel()

	alert, err = l.child.Run(ctx, db, s)
	if errors.Is(err, context.Canceled) {
		// Ignore context canceled errors
		err = nil
	}
	return alert, err

}

func (l *LimitJob) Name() string {
	return "LimitJob"
}

func NewNoopJob() *noopJob {
	return &noopJob{}
}

type noopJob struct{}

func (e *noopJob) Run(context.Context, database.DB, streaming.Sender) (*search.Alert, error) {
	return nil, nil
}

func (e *noopJob) Name() string { return "NoopJob" }
