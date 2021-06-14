package run

import (
	"context"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/search"
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

	mu      sync.Mutex
	results []result.Match
	stats   streaming.Stats
	errors  *multierror.Error
}

// Get finalises aggregation over the stream and returns the aggregated
// result. It should only be called once each do* function is finished
// running.
func (a *Aggregator) Get() ([]result.Match, streaming.Stats, *multierror.Error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.results, a.stats, a.errors
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

func (a *Aggregator) DoSymbolSearch(ctx context.Context, args *search.TextParameters, limit int) (err error) {
	tr, ctx := trace.New(ctx, "doSymbolSearch", "")
	defer func() {
		a.Error(err)
		tr.SetError(err)
		tr.Finish()
	}()

	err = SearchSymbols(ctx, args, limit, a)
	return errors.Wrap(err, "symbol search failed")
}

func (a *Aggregator) DoFilePathSearch(ctx context.Context, args *search.TextParameters) (err error) {
	tr, ctx := trace.New(ctx, "doFilePathSearch", "")
	tr.LogFields(trace.Stringer("global_search_mode", args.Mode))
	defer func() {
		a.Error(err)
		tr.SetErrorIfNotContext(err)
		tr.Finish()
	}()

	isDefaultStructuralSearch := args.PatternInfo.IsStructuralPat && args.PatternInfo.FileMatchLimit == defaultMaxSearchResults

	if !isDefaultStructuralSearch {
		return SearchFilesInRepos(ctx, args, a)
	}

	// For structural search with default limits we retry if we get no results.

	fileMatches, stats, err := SearchFilesInReposBatch(ctx, args)

	if len(fileMatches) == 0 && err == nil {
		// No results for structural search? Automatically search again and force Zoekt
		// to resolve more potential file matches by setting a higher FileMatchLimit.
		patternCopy := *(args.PatternInfo)
		patternCopy.FileMatchLimit = 1000
		argsCopy := *args
		argsCopy.PatternInfo = &patternCopy
		args = &argsCopy

		fileMatches, stats, err = SearchFilesInReposBatch(ctx, args)

		if len(fileMatches) == 0 {
			// Still no results? Give up.
			log15.Warn("Structural search gives up after more exhaustive attempt. Results may have been missed.")
			stats.IsLimitHit = false // Ensure we don't display "Show more".
		}
	}

	matches := make([]result.Match, 0, len(fileMatches))
	for _, fm := range fileMatches {
		matches = append(matches, fm)
	}

	a.Send(streaming.SearchEvent{
		Results: matches,
		Stats:   stats,
	})
	return err
}

func (a *Aggregator) DoDiffSearch(ctx context.Context, tp *search.TextParameters) (err error) {
	tr, ctx := trace.New(ctx, "doDiffSearch", "")
	defer func() {
		a.Error(err)
		tr.SetError(err)
		tr.Finish()
	}()

	if err := checkDiffCommitSearchLimits(ctx, tp, "diff"); err != nil {
		return err
	}

	args, err := ResolveCommitParameters(ctx, tp)
	if err != nil {
		log15.Warn("doDiffSearch: error while resolving commit parameters", "error", err)
		return nil
	}

	return SearchCommitDiffsInRepos(ctx, a.db, args, a)
}

func (a *Aggregator) DoCommitSearch(ctx context.Context, tp *search.TextParameters) (err error) {
	tr, ctx := trace.New(ctx, "doCommitSearch", "")
	defer func() {
		a.Error(err)
		tr.SetError(err)
		tr.Finish()
	}()

	if err := checkDiffCommitSearchLimits(ctx, tp, "commit"); err != nil {
		return err
	}

	args, err := ResolveCommitParameters(ctx, tp)
	if err != nil {
		log15.Warn("doCommitSearch: error while resolving commit parameters", "error", err)
		return nil
	}

	return SearchCommitLogInRepos(ctx, a.db, args, a)
}

func checkDiffCommitSearchLimits(ctx context.Context, args *search.TextParameters, resultType string) error {
	repos, err := getRepos(ctx, args.RepoPromise)
	if err != nil {
		return err
	}

	hasTimeFilter := false
	if _, afterPresent := args.Query.Fields()["after"]; afterPresent {
		hasTimeFilter = true
	}
	if _, beforePresent := args.Query.Fields()["before"]; beforePresent {
		hasTimeFilter = true
	}

	limits := searchrepos.SearchLimits()
	if max := limits.CommitDiffMaxRepos; !hasTimeFilter && len(repos) > max {
		return &RepoLimitError{ResultType: resultType, Max: max}
	}
	if max := limits.CommitDiffWithTimeFilterMaxRepos; hasTimeFilter && len(repos) > max {
		return &TimeLimitError{ResultType: resultType, Max: max}
	}
	return nil
}

type DiffCommitError struct {
	ResultType string
	Max        int
}

type RepoLimitError DiffCommitError
type TimeLimitError DiffCommitError

func (*RepoLimitError) Error() string {
	return "repo limit error"
}

func (*TimeLimitError) Error() string {
	return "time limit error"
}
