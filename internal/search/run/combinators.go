package run

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// NewJobWithOptional creates a combinator job from a required job and an
// optional job. When run, JobWithOptional run the required job and the
// optional job in parallel, wait for the required job to complete, then give
// the optional job a short additional amount of time (currently 100ms) before
// canceling the optional job.
func NewJobWithOptional(required Job, optional Job) Job {
	if _, ok := optional.(*noopJob); ok {
		return required
	}
	return &JobWithOptional{
		required: required,
		optional: optional,
	}
}

type JobWithOptional struct {
	required Job
	optional Job
}

func (r *JobWithOptional) Name() string {
	return fmt.Sprintf("JobWithOptional{Required: %s, Optional: %s}", r.required.Name(), r.optional.Name())
}

func (r *JobWithOptional) Run(ctx context.Context, db database.DB, s streaming.Sender) (_ *search.Alert, err error) {
	tr, ctx := trace.New(ctx, "JobWithOptional", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	start := time.Now()

	var (
		maxAlerter    search.MaxAlerter
		optionalGroup multierror.Group
		requiredGroup multierror.Group
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

	var errs *multierror.Error
	if err := requiredGroup.Wait(); err != nil {
		errs = multierror.Append(errs, err)
	}
	tr.LazyPrintf("required group completed")

	// Give optional searches some minimum budget in case required searches return quickly.
	// Cancel all remaining searches after this minimum budget.
	budget := 100 * time.Millisecond
	elapsed := time.Since(start)
	time.AfterFunc(budget-elapsed, cancel)

	if err := optionalGroup.Wait(); err != nil {
		errs = multierror.Append(errs, err)
	}
	tr.LazyPrintf("optional group completed")

	return maxAlerter.Alert, errs.ErrorOrNil()
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
	return ParallelJob(children)
}

type ParallelJob []Job

func (p ParallelJob) Name() string {
	var childNames []string
	for _, job := range p {
		childNames = append(childNames, job.Name())
	}
	return fmt.Sprintf("ParallelJob{%s}", strings.Join(childNames, ", "))
}

func (p ParallelJob) Run(ctx context.Context, db database.DB, s streaming.Sender) (_ *search.Alert, err error) {
	tr, ctx := trace.New(ctx, "ParallelJob", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	var (
		g          multierror.Group
		maxAlerter search.MaxAlerter
	)
	for _, job := range p {
		job := job
		g.Go(func() error {
			alert, err := job.Run(ctx, db, s)
			maxAlerter.Add(alert)
			return err
		})
	}
	return maxAlerter.Alert, g.Wait().ErrorOrNil()
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

func (t *TimeoutJob) Run(ctx context.Context, db database.DB, s streaming.Sender) (_ *search.Alert, err error) {
	tr, ctx := trace.New(ctx, "TimeoutJob", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	return t.child.Run(ctx, db, s)
}

func (t *TimeoutJob) Name() string {
	return fmt.Sprintf("TimeoutJob{%s}", t.child.Name())
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

func (l *LimitJob) Run(ctx context.Context, db database.DB, s streaming.Sender) (_ *search.Alert, err error) {
	tr, ctx := trace.New(ctx, "LimitJob", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	ctx, s, cancel := streaming.WithLimit(ctx, s, l.limit)
	defer cancel()

	alert, err := l.child.Run(ctx, db, s)
	if errors.Is(err, context.Canceled) {
		// Ignore context canceled errors
		err = nil
	}
	return alert, err

}

func (l *LimitJob) Name() string {
	return fmt.Sprintf("LimitJob{%s}", l.child.Name())
}

func NewNoopJob() *noopJob {
	return &noopJob{}
}

type noopJob struct{}

func (e *noopJob) Run(context.Context, database.DB, streaming.Sender) (*search.Alert, error) {
	return nil, nil
}

func (e *noopJob) Name() string { return "EmptyJob" }
