package run

import (
	"context"

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
		merger      = result.NewLiveMerger(len(a.children))
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
				stream.Send(event)
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
