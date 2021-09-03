package unindexed

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	zoektutil "github.com/sourcegraph/sourcegraph/internal/search/zoekt"
	"golang.org/x/sync/errgroup"
)

// searchRepos represent the arguments to a search called over repositories.
type searchRepos struct {
	args    *search.SearcherParameters
	repoSet *search.Repos
	stream  streaming.Sender
	indexed bool
}

// RepoFetcher is an object that exposes an interface to retrieve repos to
// search from Zoekt. The interface exposes a Get(ctx) method that allows
// parameterizing repo fetches by context.
type RepoFetcher struct {
	args              *search.TextParameters
	mode              search.GlobalSearchMode
	onMissingRepoRevs zoektutil.OnMissingRepos
}

func NewRepoFetcher(stream streaming.Sender, args *search.TextParameters) RepoFetcher {
	return RepoFetcher{
		mode:              args.Mode,
		args:              args,
		onMissingRepoRevs: zoektutil.MissingRepoRevStatus(stream),
	}
}

// Get returns the repository data to run structural search on. Importantly, it
// allows parameterizing the request to specify a context, for when multiple
// Get() calls are required with different limits or timeouts.
func (r *RepoFetcher) Get(ctx context.Context) (indexed, unindexed *search.Repos, err error) {
	request, err := zoektutil.NewIndexedSearchRequest(ctx, r.args, search.TextRequest, r.onMissingRepoRevs)
	if err != nil {
		return nil, nil, err
	}

	if r.mode == search.SearcherOnly {
		return nil, request.UnindexedRepos(), nil
	}

	return request.IndexedRepos(), request.UnindexedRepos(), nil
}

// getJob returns a function parameterized by ctx to search over repos.
func (s *searchRepos) getJob(ctx context.Context) func() error {
	return func() error {
		return callSearcherOverRepos(ctx, s.args, s.stream, s.repoSet, s.indexed)
	}
}

func runJobs(ctx context.Context, jobs []*searchRepos) error {
	g, ctx := errgroup.WithContext(ctx)
	for _, j := range jobs {
		g.Go(j.getJob(ctx))
	}
	return g.Wait()
}

// streamStructuralSearch runs structural search jobs and streams the results.
func streamStructuralSearch(ctx context.Context, args *search.SearcherParameters, repoFetcher *RepoFetcher, stream streaming.Sender) (err error) {
	ctx, stream, cleanup := streaming.WithLimit(ctx, stream, int(args.PatternInfo.FileMatchLimit))
	defer cleanup()

	indexed, unindexed, err := repoFetcher.Get(ctx)
	if err != nil {
		return err
	}

	jobs := make([]*searchRepos, 0, 2)
	searcherArgs := &search.SearcherParameters{
		SearcherURLs:    args.SearcherURLs,
		PatternInfo:     args.PatternInfo,
		UseFullDeadline: args.UseFullDeadline,
	}

	if indexed != nil {
		jobs = append(jobs, &searchRepos{args: searcherArgs, stream: stream, repoSet: indexed, indexed: true})
	}

	if unindexed != nil {
		jobs = append(jobs, &searchRepos{args: searcherArgs, stream: stream, repoSet: unindexed, indexed: false})
	}

	return runJobs(ctx, jobs)
}

// retryStructuralSearch runs a structural search with a higher limit file match
// limit so that Zoekt resolves more potential file matches.
func retryStructuralSearch(ctx context.Context, args *search.SearcherParameters, repoFetcher *RepoFetcher, stream streaming.Sender) error {
	patternCopy := *(args.PatternInfo)
	patternCopy.FileMatchLimit = 1000
	argsCopy := *args
	argsCopy.PatternInfo = &patternCopy
	args = &argsCopy
	return streamStructuralSearch(ctx, args, repoFetcher, stream)
}

func runStructuralSearch(ctx context.Context, args *search.SearcherParameters, repoFetcher *RepoFetcher, stream streaming.Sender) error {
	if args.PatternInfo.FileMatchLimit != search.DefaultMaxSearchResults {
		// streamStructuralSearch performs a streaming search when the user sets a value
		// for `count`. The first return parameter indicates whether the request was
		// serviced with streaming.
		return streamStructuralSearch(ctx, args, repoFetcher, stream)
	}

	// For structural search with default limits we retry if we get no results.
	fileMatches, stats, err := streaming.CollectStream(func(stream streaming.Sender) error {
		return streamStructuralSearch(ctx, args, repoFetcher, stream)
	})

	if len(fileMatches) == 0 && err == nil {
		// retry structural search with a higher limit.
		fileMatches, stats, err = streaming.CollectStream(func(stream streaming.Sender) error {
			return retryStructuralSearch(ctx, args, repoFetcher, stream)
		})
		if err != nil {
			return err
		}

		if len(fileMatches) == 0 {
			// Still no results? Give up.
			log15.Warn("Structural search gives up after more exhaustive attempt. Results may have been missed.")
			stats.IsLimitHit = false // Ensure we don't display "Show more".
		}
	}

	matches := make([]result.Match, 0, len(fileMatches))
	for _, fm := range fileMatches {
		if _, ok := fm.(*result.FileMatch); !ok {
			return errors.Errorf("StructuralSearch failed to convert results")
		}
		matches = append(matches, fm)
	}

	stream.Send(streaming.SearchEvent{
		Results: matches,
		Stats:   stats,
	})
	return err
}

type StructuralSearch struct {
	RepoFetcher  RepoFetcher
	Mode         search.GlobalSearchMode
	SearcherArgs search.SearcherParameters
}

func (s *StructuralSearch) Run(ctx context.Context, stream streaming.Sender) error {
	return runStructuralSearch(ctx, &s.SearcherArgs, &s.RepoFetcher, stream)
}
