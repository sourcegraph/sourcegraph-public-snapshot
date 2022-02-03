package graphqlbackend

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/sync/semaphore"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// NewAndJob creates a job that will run each of its child jobs and only
// stream matches that were found in all of the child jobs.
func NewAndJob(children ...run.Job) run.Job {
	if len(children) == 0 {
		return run.NewNoopJob()
	} else if len(children) == 1 {
		return children[0]
	}
	return &AndJob{children: children}
}

type AndJob struct {
	children []run.Job
}

func (a *AndJob) Run(ctx context.Context, db database.DB, stream streaming.Sender) (_ *search.Alert, err error) {
	tr, ctx := trace.New(ctx, "AndJob", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	var (
		g          multierror.Group
		maxAlerter search.MaxAlerter
		sem        = semaphore.NewWeighted(16)
		merger     = result.NewLiveMerger(len(a.children))
	)
	for childNum, child := range a.children {
		childNum, child := childNum, child
		g.Go(func() error {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}
			defer sem.Release(1)

			intersectingStream := streaming.StreamFunc(func(event streaming.SearchEvent) {
				event.Results = merger.AddMatches(event.Results, childNum)
				stream.Send(event)
			})

			alert, err := child.Run(ctx, db, intersectingStream)
			maxAlerter.Add(alert)
			return err
		})
	}

	return maxAlerter.Alert, g.Wait().ErrorOrNil()
}

func (a *AndJob) Name() string {
	return "AndJob"
}
