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

// StructuralSearchFilesInRepos searches a set of repos for a structural pattern.
func StructuralSearchFilesInRepos(ctx context.Context, args *search.TextParameters, stream streaming.Sender) (err error) {
	ctx, stream, cleanup := streaming.WithLimit(ctx, stream, int(args.PatternInfo.FileMatchLimit))
	defer cleanup()

	indexed, err := textSearchRequest(ctx, args, zoektutil.MissingRepoRevStatus(stream))
	if err != nil {
		return err
	}

	jobs := []withContext{}
	if args.Mode != search.SearcherOnly {
		// Job for indexed repositories (fulfilled via searcher).
		jobs = append(jobs, structuralSearchJob(args, stream, IndexedMap(indexed.Repos())))
	}
	// Job for unindexed repositories.
	jobs = append(jobs, structuralSearchJob(args, stream, UnindexedList(indexed.Unindexed)))

	g, ctx := errgroup.WithContext(ctx)
	for _, job := range jobs {
		g.Go(job(ctx))
	}

	return g.Wait()
}

func StructuralSearch(ctx context.Context, args *search.TextParameters, stream streaming.Sender) error {
	if args.PatternInfo.FileMatchLimit != search.DefaultMaxSearchResults {
		// Service structural search via SearchFilesInRepos when we have
		// an explicit `count` value that differs from the default value
		// (e.g., user sets higher counts).
		return StructuralSearchFilesInRepos(ctx, args, stream)
	}

	// For structural search with default limits we retry if we get no results.
	fileMatches, stats, err := streaming.CollectStream(func(stream streaming.Sender) error {
		return StructuralSearchFilesInRepos(ctx, args, stream)
	})
	if err != nil {
		return err
	}

	if len(fileMatches) == 0 {
		// No results for structural search? Automatically search again and force Zoekt
		// to resolve more potential file matches by setting a higher FileMatchLimit.
		patternCopy := *(args.PatternInfo)
		patternCopy.FileMatchLimit = 1000
		argsCopy := *args
		argsCopy.PatternInfo = &patternCopy
		args = &argsCopy

		fileMatches, stats, err = streaming.CollectStream(func(stream streaming.Sender) error {
			return StructuralSearchFilesInRepos(ctx, args, stream)
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
