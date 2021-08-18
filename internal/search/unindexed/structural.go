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

// repoData represents an object of repository revisions to search.
type repoData interface {
	AsList() []*search.RepositoryRevisions
	IsIndexed() bool
}

type IndexedMap map[string]*search.RepositoryRevisions

func (m IndexedMap) AsList() []*search.RepositoryRevisions {
	reposList := make([]*search.RepositoryRevisions, 0, len(m))
	for _, repo := range m {
		reposList = append(reposList, repo)
	}
	return reposList
}

func (IndexedMap) IsIndexed() bool {
	return true
}

type UnindexedList []*search.RepositoryRevisions

func (ul UnindexedList) AsList() []*search.RepositoryRevisions {
	return ul
}

func (UnindexedList) IsIndexed() bool {
	return false
}

// The following type definitions compose separable concerns for running structural search.

// searchJob is a function that may run in its own Go routine.
type searchJob func() error

// withContext parameterizes a searchJob by context, making it easy to add ctx for multiple jobs part of an errgroup.
type withContext func(context.Context) searchJob

// structuralSearchJob creates a composable function for running structural
// search. It unrolls the context parameter so that multiple jobs can be
// parameterized by the errgroup context.
func structuralSearchJob(args *search.TextParameters, stream streaming.Sender, repoData repoData) withContext {
	return func(ctx context.Context) searchJob {
		return func() error {
			return callSearcherOverRepos(ctx, args, stream, repoData.AsList(), repoData.IsIndexed())
		}
	}
}

func runJobs(ctx context.Context, jobs []withContext) error {
	g, ctx := errgroup.WithContext(ctx)
	for _, job := range jobs {
		g.Go(job(ctx))
	}
	return g.Wait()
}

// repoSets returns the set of repositories to search (whether indexed or unindexed) based on search mode.
func repoSets(request zoektutil.IndexedSearchRequest, mode search.GlobalSearchMode) []repoData {
	repoSets := []repoData{UnindexedList(request.UnindexedRepos())} // unindexed included by default
	if mode != search.SearcherOnly {
		repoSets = append(repoSets, IndexedMap(request.IndexedRepos()))
	}
	return repoSets
}

// streamStructuralSearch runs structural search jobs and streams the results.
func streamStructuralSearch(ctx context.Context, args *search.TextParameters, stream streaming.Sender) (err error) {
	ctx, stream, cleanup := streaming.WithLimit(ctx, stream, int(args.PatternInfo.FileMatchLimit))
	defer cleanup()

	request, err := textSearchRequest(ctx, args, zoektutil.MissingRepoRevStatus(stream))
	if err != nil {
		return err
	}

	jobs := []withContext{}
	for _, repoSet := range repoSets(request, args.Mode) {
		jobs = append(jobs, structuralSearchJob(args, stream, repoSet))
	}
	return runJobs(ctx, jobs)
}

// retryStructuralSearch runs a structural search with a higher limit file match
// limit so that Zoekt resolves more potential file matches.
func retryStructuralSearch(ctx context.Context, args *search.TextParameters, stream streaming.Sender) error {
	patternCopy := *(args.PatternInfo)
	patternCopy.FileMatchLimit = 1000
	argsCopy := *args
	argsCopy.PatternInfo = &patternCopy
	args = &argsCopy
	return streamStructuralSearch(ctx, args, stream)
}

func StructuralSearch(ctx context.Context, args *search.TextParameters, stream streaming.Sender) error {
	if args.PatternInfo.FileMatchLimit != search.DefaultMaxSearchResults {
		// streamStructuralSearch performs a streaming search when the user sets a value
		// for `count`. The first return parameter indicates whether the request was
		// serviced with streaming.
		return streamStructuralSearch(ctx, args, stream)
	}

	// For structural search with default limits we retry if we get no results.
	fileMatches, stats, err := streaming.CollectStream(func(stream streaming.Sender) error {
		return streamStructuralSearch(ctx, args, stream)
	})

	if len(fileMatches) == 0 && err == nil {
		// retry structural search with a higher limit.
		fileMatches, stats, err = streaming.CollectStream(func(stream streaming.Sender) error {
			return retryStructuralSearch(ctx, args, stream)
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
