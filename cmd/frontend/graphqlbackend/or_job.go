package graphqlbackend

import (
	"context"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/sync/semaphore"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type OrJob struct {
	limit    int
	children []run.Job
}

func (j *OrJob) Run(ctx context.Context, db database.DB, stream streaming.Sender) (_ *search.Alert, err error) {
	tr, ctx := trace.New(ctx, "OrJob", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	var (
		// NOTE(tsenart): In the future, when we have the need for more intelligent rate limiting,
		// this concurrency limit should probably be informed by a user's rate limit quota
		// at any given time.
		sem        = semaphore.NewWeighted(16)
		maxAlerter search.MaxAlerter
		g          multierror.Group

		mu        sync.Mutex
		dedup     = result.NewDeduper()
		stats     streaming.Stats
		remaining = j.limit
	)

	for _, child := range j.children {
		child := child
		g.Go(func() error {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}
			defer sem.Release(1)

			agg := streaming.NewAggregatingStream()
			alert, err := child.Run(ctx, db, agg)
			maxAlerter.Add(alert)
			if err != nil {
				return err
			}

			mu.Lock()
			defer mu.Unlock()

			if remaining <= 0 {
				return context.Canceled
			}

			// BUG: When we find enough results we stop adding them to dedupper,
			// but don't adjust the stats accordingly. This bug was here
			// before, and remains after making OR query evaluation concurrent.
			stats.Update(&agg.Stats)

			for _, m := range agg.Results {
				remaining = m.Limit(remaining)
				if dedup.Add(m); remaining <= 0 {
					return context.Canceled
				}
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return maxAlerter.Alert, err
	}

	stream.Send(streaming.SearchEvent{
		Results: dedup.Results(),
		Stats:   stats,
	})
	return maxAlerter.Alert, nil
}

func (j *OrJob) Name() string {
	return "OrJob"
}
