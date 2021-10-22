package run

import (
	"context"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/commit"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/search/symbol"
	"github.com/sourcegraph/sourcegraph/internal/search/unindexed"
	zoektutil "github.com/sourcegraph/sourcegraph/internal/search/zoekt"
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

func (a *Aggregator) Send(event streaming.SearchEvent) {
	if a.parentStream != nil {
		a.parentStream.Send(event)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Do not aggregate results if we are streaming.
	if a.parentStream == nil {
		a.results = append(a.results, event.Results...)
	}

	a.matchCount += len(event.Results)
	a.stats.Update(&event.Stats)
}

func (a *Aggregator) Error(err error) {
	a.mu.Lock()
	a.errors = multierror.Append(a.errors, err)
	a.mu.Unlock()
}

func (a *Aggregator) DoRepoSearch(ctx context.Context, args *search.TextParameters, limit int32) (err error) {
	tr, ctx := trace.New(ctx, "doRepoSearch", "")
	defer func() {
		a.Error(err)
		tr.SetError(err)
		tr.Finish()
	}()

	err = SearchRepositories(ctx, args, limit, a)
	return errors.Wrap(err, "repository search failed")
}

func (a *Aggregator) DoSearch(ctx context.Context, job Job, mode search.GlobalSearchMode) (err error) {
	tr, ctx := trace.New(ctx, "DoSearch", job.Name())
	tr.LogFields(trace.Stringer("global_search_mode", mode))
	defer func() {
		a.Error(err)
		tr.SetErrorIfNotContext(err)
		tr.Finish()
	}()

	err = job.Run(ctx, a)
	return errors.Wrap(err, job.Name()+" search failed")

}

func (a *Aggregator) DoSymbolSearch(ctx context.Context, args *search.TextParameters, limit int) (err error) {
	tr, ctx := trace.New(ctx, "doSymbolSearch", "")
	defer func() {
		a.Error(err)
		tr.SetError(err)
		tr.Finish()
	}()

	err = symbol.Search(ctx, args, limit, a)
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

func (a *Aggregator) DoDiffSearch(ctx context.Context, tp *search.TextParameters) (err error) {
	tr, ctx := trace.New(ctx, "doDiffSearch", "")
	defer func() {
		a.Error(err)
		tr.SetError(err)
		tr.Finish()
	}()

	if err := commit.CheckSearchLimits(tp.Query, len(tp.Repos), "diff"); err != nil {
		return err
	}

	args, err := commit.ResolveCommitParameters(ctx, tp)
	if err != nil {
		log15.Warn("doDiffSearch: error while resolving commit parameters", "error", err)
		return nil
	}

	return commit.SearchCommitDiffsInRepos(ctx, a.db, args, a)
}

func (a *Aggregator) DoCommitSearch(ctx context.Context, tp *search.TextParameters) (err error) {
	tr, ctx := trace.New(ctx, "doCommitSearch", "")
	defer func() {
		a.Error(err)
		tr.SetError(err)
		tr.Finish()
	}()

	if err := commit.CheckSearchLimits(tp.Query, len(tp.Repos), "commit"); err != nil {
		return err
	}

	args, err := commit.ResolveCommitParameters(ctx, tp)
	if err != nil {
		log15.Warn("doCommitSearch: error while resolving commit parameters", "error", err)
		return nil
	}

	return commit.SearchCommitLogInRepos(ctx, a.db, args, a)
}
