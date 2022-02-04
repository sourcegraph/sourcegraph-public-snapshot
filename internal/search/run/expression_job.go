package run

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"
	"go.uber.org/atomic"
	"golang.org/x/sync/semaphore"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// NewAndJob creates a job that will run each of its child jobs and only
// stream matches that were found in all of the child jobs.
func NewAndJob(children ...Job) Job {
	if len(children) == 0 {
		return NewNoopJob()
	} else if len(children) == 1 {
		return children[0]
	}
	return &AndJob{children: children}
}

type AndJob struct {
	children []Job
}

func (a *AndJob) Run(ctx context.Context, db database.DB, stream streaming.Sender) (_ *search.Alert, err error) {
	tr, ctx := trace.New(ctx, "AndJob", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	var (
		g           multierror.Group
		maxAlerter  search.MaxAlerter
		limitHit    atomic.Bool
		sentResults atomic.Bool
		sem         = semaphore.NewWeighted(16)
		merger      = result.NewMerger(len(a.children))
	)
	for childNum, child := range a.children {
		childNum, child := childNum, child
		g.Go(func() error {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}
			defer sem.Release(1)

			intersectingStream := streaming.StreamFunc(func(event streaming.SearchEvent) {
				if event.Stats.IsLimitHit {
					limitHit.Store(true)
				}
				event.Results = merger.AddMatches(event.Results, childNum)
				if len(event.Results) > 0 {
					sentResults.Store(true)
				}
				if len(event.Results) > 0 || !event.Stats.Zero() {
					stream.Send(event)
				}
			})

			alert, err := child.Run(ctx, db, intersectingStream)
			maxAlerter.Add(alert)
			return err
		})
	}

	err = g.Wait().ErrorOrNil()

	if !sentResults.Load() && limitHit.Load() {
		maxAlerter.Add(search.AlertForCappedAndExpression())
	}
	return maxAlerter.Alert, g.Wait().ErrorOrNil()
}

func (a *AndJob) Name() string {
	return "AndJob"
}

// NewAndJob creates a job that will run each of its child jobs and stream
// deduplicated matches that were streamed by at least one of the jobs.
func NewOrJob(children ...Job) Job {
	return &OrJob{
		children: children,
	}
}

type OrJob struct {
	children []Job
}

func (j *OrJob) Run(ctx context.Context, db database.DB, stream streaming.Sender) (_ *search.Alert, err error) {
	tr, ctx := trace.New(ctx, "OrJob", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	var (
		maxAlerter search.MaxAlerter
		g          multierror.Group
		sem        = semaphore.NewWeighted(16)
		merger     = result.NewMerger(len(j.children))
	)
	for childNum, child := range j.children {
		childNum, child := childNum, child
		g.Go(func() error {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}
			defer sem.Release(1)

			unioningStream := streaming.StreamFunc(func(event streaming.SearchEvent) {
				event.Results = merger.AddMatches(event.Results, childNum)
				if len(event.Results) > 0 || !event.Stats.Zero() {
					stream.Send(event)
				}
			})

			alert, err := child.Run(ctx, db, unioningStream)
			maxAlerter.Add(alert)
			return err
		})
	}

	// TODO(@camdencheek): errors.Is isn't good enough here since a single
	// backend that returns a context.Canceled error will make the multierror
	// return true for errors.Is(err, context.Canceled). Ideally, we have some
	// sort of multi-error filter that can filter out any context.Canceled and
	// leave us with whatever errors are left. Note that this is true of anywhere
	// we check the type of an aggregated error. This is neither a new nor a
	// unique problem.
	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return maxAlerter.Alert, err
	}

	// Send results that were only seen by some of the sources
	stream.Send(streaming.SearchEvent{
		Results: merger.UnsentTracked(),
	})
	return maxAlerter.Alert, nil
}

func (j *OrJob) Name() string {
	return "OrJob"
}
