package run

import (
	"context"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/search"
	searchrepos "github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/search/symbol"
	"github.com/sourcegraph/sourcegraph/internal/search/unindexed"
	zoektutil "github.com/sourcegraph/sourcegraph/internal/search/zoekt"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func NewAggregator(db dbutil.DB, stream streaming.Sender, filterFunc func(*streaming.SearchEvent) error) *Aggregator {
	return &Aggregator{
		db:           db,
		parentStream: stream,
		filterFunc:   filterFunc,
		errors:       &multierror.Error{},
	}
}

type Aggregator struct {
	parentStream streaming.Sender
	db           dbutil.DB

	// filterFunc can be applied to manipulate each SearchEvent before it gets propagated.
	// It is currently used to provide sub-repo perms filtering.
	//
	// SearchEvent is still propagated even in an error case - filterFunc should make sure
	// the appropriate manipulations are made before returning an error.
	filterFunc func(*streaming.SearchEvent) error

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
	if a.filterFunc != nil {
		// We don't need to return if we encounter an error because filterFunc should
		// remove anything that should not be provided to the user before returning.
		if err := a.filterFunc(&event); err != nil {
			a.Error(err)
		}
	}

	if a.parentStream != nil {
		a.parentStream.Send(event)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Only aggregate results if we are not streaming.
	if a.parentStream == nil {
		a.results = append(a.results, event.Results...)

		if a.stats.Repos == nil {
			a.stats.Repos = make(map[api.RepoID]types.MinimalRepo)
		}

		for _, r := range event.Results {
			repo := r.RepoName()
			if _, ok := a.stats.Repos[repo.ID]; !ok {
				a.stats.Repos[repo.ID] = repo
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

func (a *Aggregator) DoSearch(ctx context.Context, job Job, repos searchrepos.Pager, mode search.GlobalSearchMode) (err error) {
	tr, ctx := trace.New(ctx, "DoSearch", job.Name())
	tr.LogFields(trace.Stringer("global_search_mode", mode))
	defer func() {
		a.Error(err)
		tr.SetErrorIfNotContext(err)
		tr.Finish()
	}()

	err = job.Run(ctx, a, repos)
	return errors.Wrap(err, job.Name()+" search failed")
}

func (a *Aggregator) DoSymbolSearch(ctx context.Context, args *search.TextParameters, notSearcherOnly, globalSearch bool, limit int) (err error) {
	tr, ctx := trace.New(ctx, "doSymbolSearch", "")
	defer func() {
		a.Error(err)
		tr.SetError(err)
		tr.Finish()
	}()

	err = symbol.Search(ctx, args, notSearcherOnly, globalSearch, limit, a)
	return errors.Wrap(err, "symbol search failed")
}

func (a *Aggregator) DoFilePathSearch(ctx context.Context, zoektArgs zoektutil.IndexedSearchRequest, searcherArgs *search.SearcherParameters, notSearcherOnly bool, stream streaming.Sender) (err error) {
	tr, ctx := trace.New(ctx, "doFilePathSearch", "")
	defer func() {
		a.Error(err)
		tr.SetErrorIfNotContext(err)
		tr.Finish()
	}()

	return unindexed.SearchFilesInRepos(ctx, zoektArgs, searcherArgs, notSearcherOnly, stream)
}
