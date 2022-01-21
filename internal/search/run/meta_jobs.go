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

type RequiredAndOptionalJob struct {
	Required Job
	Optional Job
}

func (r *RequiredAndOptionalJob) Name() string {
	return fmt.Sprintf("RequiredAndOptionalJob{Required: %s, Optional: %s}", r.Required.Name(), r.Optional.Name())
}

func (r *RequiredAndOptionalJob) Run(ctx context.Context, s streaming.Sender, pager searchrepos.Pager) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var optionalGroup, requiredGroup multierror.Group
	requiredGroup.Go(func() error {
		return r.Required.Run(ctx, s, pager)
	})
	optionalGroup.Go(func() error {
		return r.Optional.Run(ctx, s, pager)
	})

	var errs *multierror.Error
	if err := requiredGroup.Wait(); err != nil {
		errs = multierror.Append(errs, err)
	}

	// Give optional searches some minimum budget in case required searches return quickly.
	// Cancel all remaining searches after this minimum budget.
	budget := 100 * time.Millisecond
	time.AfterFunc(budget, cancel)

	if err := optionalGroup.Wait(); err != nil {
		errs = multierror.Append(errs, err)
	}

	return errs.ErrorOrNil()
}

type ParallelJob struct {
	Jobs []Job
}

func (p *ParallelJob) Name() string {
	var childNames []string
	for _, job := range p.Jobs {
		childNames = append(childNames, job.Name())
	}
	return fmt.Sprintf("ParallelJob{%s}", strings.Join(childNames, ", "))
}

func (p *ParallelJob) Run(ctx context.Context, s streaming.Sender, pager searchrepos.Pager) error {
	var g multierror.Group
	for _, job := range p.Jobs {
		job := job
		g.Go(func() error {
			return job.Run(ctx, s, pager)
		})
	}
	return g.Wait()
}
