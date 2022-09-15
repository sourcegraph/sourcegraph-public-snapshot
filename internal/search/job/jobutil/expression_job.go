package jobutil

import (
	"context"

	"github.com/opentracing/opentracing-go/log"
	"go.uber.org/atomic"
	"golang.org/x/sync/semaphore"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewAndJob creates a job that will run each of its child jobs and only
// stream matches that were found in all of the child jobs.
func NewAndJob(children ...job.Job) job.Job {
	if len(children) == 0 {
		return NewNoopJob()
	} else if len(children) == 1 {
		return children[0]
	}
	return &AndJob{children: children}
}

type AndJob struct {
	children []job.Job
}

func (a *AndJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, stream, a)
	defer func() { finish(alert, err) }()

	var (
		g           errors.Group
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

			alert, err := child.Run(ctx, clients, intersectingStream)
			maxAlerter.Add(alert)
			return err
		})
	}

	err = g.Wait()

	if !sentResults.Load() && limitHit.Load() {
		maxAlerter.Add(search.AlertForCappedAndExpression())
	}
	return maxAlerter.Alert, g.Wait()
}

func (a *AndJob) Name() string {
	return "AndJob"
}

func (a *AndJob) Fields(job.Verbosity) []log.Field { return nil }

func (a *AndJob) Children() []job.Describer {
	res := make([]job.Describer, len(a.children))
	for i := range a.children {
		res[i] = a.children[i]
	}
	return res
}

func (a *AndJob) MapChildren(fn job.MapFunc) job.Job {
	cp := *a
	cp.children = make([]job.Job, len(a.children))
	for i := range a.children {
		cp.children[i] = job.Map(a.children[i], fn)
	}
	return &cp
}

// NewAndJob creates a job that will run each of its child jobs and stream
// deduplicated matches that were streamed by at least one of the jobs.
func NewOrJob(children ...job.Job) job.Job {
	if len(children) == 0 {
		return NewNoopJob()
	} else if len(children) == 1 {
		return children[0]
	}
	return &OrJob{
		children: children,
	}
}

type OrJob struct {
	children []job.Job
}

// For OR queries, there are two phases:
//  1. Stream any results that are found in every subquery
//  2. Once all subqueries have completed, send the results we've found that
//     were returned by some subqueries, but not all subqueries.
//
// This means that the only time we would hit streaming limit before we have
// results from all subqueries is if we hit the limit only with results from
// phase 1. These results are very "fair" in that they are found in all
// subqueries.
//
// Then, in phase 2, we send all results that were returned by at least one
// sub-query. These are generated from a map iteration, so the document order
// is random, meaning that when/if they are truncated to fit inside the limit,
// they will be from a random distribution of sub-queries.
//
// This solution has the following nice properties:
//   - Early cancellation is possible
//   - Results are streamed where possible, decreasing user-visible latency
//   - The only results that are streamed are "fair" results. They are "fair" because
//     they were returned from every subquery, so there can be no bias between subqueries
//   - The only time we cancel early is when streamed results hit the limit. Since the only
//     streamed results are "fair" results, there will be no bias against slow or low-volume subqueries
//   - Every result we stream is guaranteed to be "complete". By "complete", I mean if I search for "a or b",
//     the streamed result will highlight both "a" and "b" if they both exist in the document.
//   - The bias is towards documents that match all of our subqueries, so doesn't bias any individual subquery.
//     Additionally, a bias towards matching all subqueries is probably desirable, since it's more likely that
//     a document matching all subqueries is what the user is looking for than a document matching only one.
func (j *OrJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, stream, j)
	defer func() { finish(alert, err) }()

	var (
		maxAlerter search.MaxAlerter
		g          errors.Group
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

			alert, err := child.Run(ctx, clients, unioningStream)
			maxAlerter.Add(alert)
			return err
		})
	}

	err = g.Wait()

	// Send results that were only seen by some of the sources, regardless of
	// whether we got an error from any of our children.
	unsentTracked := merger.UnsentTracked()
	if len(unsentTracked) > 0 {
		stream.Send(streaming.SearchEvent{
			Results: unsentTracked,
		})
	}

	return maxAlerter.Alert, errors.Ignore(err, errors.IsContextCanceled)
}

func (j *OrJob) Name() string {
	return "OrJob"
}

func (j *OrJob) Fields(job.Verbosity) []log.Field { return nil }

func (j *OrJob) Children() []job.Describer {
	res := make([]job.Describer, len(j.children))
	for i := range j.children {
		res[i] = j.children[i]
	}
	return res
}

func (j *OrJob) MapChildren(fn job.MapFunc) job.Job {
	cp := *j
	cp.children = make([]job.Job, len(j.children))
	for i := range j.children {
		cp.children[i] = job.Map(j.children[i], fn)
	}
	return &cp
}
