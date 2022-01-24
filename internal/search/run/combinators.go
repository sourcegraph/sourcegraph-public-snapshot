package run

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

// NewJobWithOptional creates a combinator job from a required job and an
// optional job. When run, JobWithOptional run the required job and the
// optional job in parallel, wait for the required job to complete, then give
// the optional job a short additional amount of time (currently 100ms) before
// canceling the optional job.
func NewJobWithOptional(required Job, optional Job) Job {
	if _, ok := optional.(*emptyJob); ok {
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

func (r *JobWithOptional) Run(ctx context.Context, db database.DB, s streaming.Sender) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	start := time.Now()

	var optionalGroup, requiredGroup multierror.Group
	requiredGroup.Go(func() error {
		return r.required.Run(ctx, db, s)
	})
	optionalGroup.Go(func() error {
		return r.optional.Run(ctx, db, s)
	})

	var errs *multierror.Error
	if err := requiredGroup.Wait(); err != nil {
		errs = multierror.Append(errs, err)
	}

	// Give optional searches some minimum budget in case required searches return quickly.
	// Cancel all remaining searches after this minimum budget.
	budget := 100 * time.Millisecond
	elapsed := time.Since(start)
	time.AfterFunc(budget-elapsed, cancel)

	if err := optionalGroup.Wait(); err != nil {
		errs = multierror.Append(errs, err)
	}

	return errs.ErrorOrNil()
}

// NewParallelJob will create a job that runs all its child jobs in separate
// goroutines, then waits for all to complete. It returns an aggregated error
// if any of the child jobs failed.
func NewParallelJob(children ...Job) Job {
	if len(children) == 0 {
		return &emptyJob{}
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

func (p ParallelJob) Run(ctx context.Context, db database.DB, s streaming.Sender) error {
	var g multierror.Group
	for _, job := range p {
		job := job
		g.Go(func() error {
			return job.Run(ctx, db, s)
		})
	}
	return g.Wait().ErrorOrNil()
}

type emptyJob struct{}

func (e *emptyJob) Run(context.Context, database.DB, streaming.Sender) error { return nil }
func (e *emptyJob) Name() string                                             { return "EmptyJob" }
