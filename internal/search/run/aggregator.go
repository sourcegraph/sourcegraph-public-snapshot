package run

import (
	"context"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	searchrepos "github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func NewAggregator(db dbutil.DB, stream streaming.Sender) *Aggregator {
	return &Aggregator{
		db:           db,
		parentStream: stream,
		errors:       &multierror.Error{},
	}
}

type Aggregator struct {
	parentStream streaming.Sender
	db           dbutil.DB

	mu         sync.Mutex
	results    []result.Match
	stats      streaming.Stats
	errors     *multierror.Error
	matchCount int
}

// Get finalises aggregation over the stream and returns the aggregated
// result. It should only be called once each do* function is finished
// running.
func (a *Aggregator) Get() ([]result.Match, streaming.Stats, int, *multierror.Error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.results, a.stats, a.matchCount, a.errors
}

// Send propagates the given event to the Aggregator's parent stream, or
// aggregates it within results.
//
// It currently also applies sub-repo permissions filtering (see inline docs).
func (a *Aggregator) Send(event streaming.SearchEvent) {
	if a.parentStream != nil {
		a.parentStream.Send(event)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Only aggregate results if we are not streaming.
	if a.parentStream == nil {
		a.results = append(a.results, event.Results...)

		if a.stats.Repos == nil {
			a.stats.Repos = make(map[api.RepoID]struct{})
		}

		for _, r := range event.Results {
			repo := r.RepoName()
			if _, ok := a.stats.Repos[repo.ID]; !ok {
				a.stats.Repos[repo.ID] = struct{}{}
			}
		}
	}

	a.matchCount += len(event.Results)
	a.stats.Update(&event.Stats)
}

func (a *Aggregator) Error(err error) {
	a.mu.Lock()
	a.errors = multierror.Append(a.errors, err)
	a.mu.Unlock()
	if err != nil {
		log15.Debug("aggregated search error", "error", err)
	}
}

func (a *Aggregator) DoSearch(ctx context.Context, job Job, repos searchrepos.Pager) (err error) {
	tr, ctx := trace.New(ctx, "DoSearch", job.Name())
	defer func() {
		a.Error(err)
		tr.SetErrorIfNotContext(err)
		tr.Finish()
	}()

	err = job.Run(ctx, a, repos)
	return errors.Wrap(err, job.Name()+" search failed")
}
