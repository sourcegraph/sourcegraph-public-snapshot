package run

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"

	searchrepos "github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

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
	return fmt.Sprintf("RequiredAndOptionalJob{Required: %s, Optional: %s}", r.required.Name(), r.optional.Name())
}

func (r *JobWithOptional) Run(ctx context.Context, s streaming.Sender, pager searchrepos.Pager) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	start := time.Now()

	var optionalGroup, requiredGroup multierror.Group
	requiredGroup.Go(func() error {
		return r.required.Run(ctx, s, pager)
	})
	optionalGroup.Go(func() error {
		return r.optional.Run(ctx, s, pager)
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

func NewParallelJob(children ...Job) Job {
	if len(children) == 0 {
		return &emptyJob{}
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
	var childNames []string
	for _, job := range p.children {
		childNames = append(childNames, job.Name())
	}
	return fmt.Sprintf("ParallelJob{%s}", strings.Join(childNames, ", "))
}

func (p *ParallelJob) Run(ctx context.Context, s streaming.Sender, pager searchrepos.Pager) error {
	var g multierror.Group
	for _, job := range p.children {
		job := job
		g.Go(func() error {
			return job.Run(ctx, s, pager)
		})
	}
	return g.Wait()
}

type emptyJob struct{}

func (e *emptyJob) Run(context.Context, streaming.Sender, searchrepos.Pager) error { return nil }
func (e *emptyJob) Name() string                                                   { return "EmptyJob" }
