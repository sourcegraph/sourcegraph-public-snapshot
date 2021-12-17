package unindexed

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	searchrepos "github.com/sourcegraph/sourcegraph/internal/search/repos"
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

type IndexedMap map[api.RepoID]*search.RepositoryRevisions

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

// searchRepos represent the arguments to a search called over repositories.
type searchRepos struct {
	args    *search.SearcherParameters
	repoSet repoData
	stream  streaming.Sender
}

func PartitionRepos(request zoektutil.IndexedSearchRequest, notSearcherOnly bool) ([]repoData, error) {
	repoSets := []repoData{UnindexedList(request.UnindexedRepos())} // unindexed included by default
	if notSearcherOnly {
		repoSets = append(repoSets, IndexedMap(request.IndexedRepos()))
	}
	return repoSets, nil
}

// getJob returns a function parameterized by ctx to search over repos.
func (s *searchRepos) getJob(ctx context.Context) func() error {
	return func() error {
		return callSearcherOverRepos(ctx, s.args, s.stream, s.repoSet.AsList(), s.repoSet.IsIndexed())
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
func streamStructuralSearch(ctx context.Context, args *search.SearcherParameters, repos []repoData, stream streaming.Sender) (err error) {
	ctx, stream, cleanup := streaming.WithLimit(ctx, stream, int(args.PatternInfo.FileMatchLimit))
	defer cleanup()

	jobs := []*searchRepos{}
	for _, repoSet := range repos {
		searcherArgs := &search.SearcherParameters{
			SearcherURLs:    args.SearcherURLs,
			PatternInfo:     args.PatternInfo,
			UseFullDeadline: args.UseFullDeadline,
		}

		jobs = append(jobs, &searchRepos{args: searcherArgs, stream: stream, repoSet: repoSet})
	}
	return runJobs(ctx, jobs)
}

// retryStructuralSearch runs a structural search with a higher limit file match
// limit so that Zoekt resolves more potential file matches.
func retryStructuralSearch(ctx context.Context, args *search.SearcherParameters, repos []repoData, stream streaming.Sender) error {
	patternCopy := *(args.PatternInfo)
	patternCopy.FileMatchLimit = 1000
	argsCopy := *args
	argsCopy.PatternInfo = &patternCopy
	args = &argsCopy
	return streamStructuralSearch(ctx, args, repos, stream)
}

func runStructuralSearch(ctx context.Context, args *search.SearcherParameters, repos []repoData, stream streaming.Sender) error {
	if args.PatternInfo.FileMatchLimit != search.DefaultMaxSearchResults {
		// streamStructuralSearch performs a streaming search when the user sets a value
		// for `count`. The first return parameter indicates whether the request was
		// serviced with streaming.
		return streamStructuralSearch(ctx, args, repos, stream)
	}

	// For structural search with default limits we retry if we get no results.
	fileMatches, stats, err := streaming.CollectStream(func(stream streaming.Sender) error {
		return streamStructuralSearch(ctx, args, repos, stream)
	})

	if len(fileMatches) == 0 && err == nil {
		// retry structural search with a higher limit.
		fileMatches, stats, err = streaming.CollectStream(func(stream streaming.Sender) error {
			return retryStructuralSearch(ctx, args, repos, stream)
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
	ZoektArgs    *search.ZoektParameters
	SearcherArgs *search.SearcherParameters

	NotSearcherOnly   bool
	UseIndex          query.YesNoOnly
	ContainsRefGlobs  bool
	OnMissingRepoRevs zoektutil.OnMissingRepoRevs

	IsRequired bool
}

func (s *StructuralSearch) Run(ctx context.Context, stream streaming.Sender, repos searchrepos.Pager) error {
	return repos.Paginate(ctx, nil, func(page *searchrepos.Resolved) error {
		request, ok, err := zoektutil.OnlyUnindexed(page.RepoRevs, s.ZoektArgs.Zoekt, s.UseIndex, s.ContainsRefGlobs, s.OnMissingRepoRevs)
		if err != nil {
			return err
		}
		if !ok {
			request, err = zoektutil.NewIndexedSubsetSearchRequest(ctx, page.RepoRevs, s.UseIndex, s.ZoektArgs, s.OnMissingRepoRevs)
			if err != nil {
				return err
			}
		}

		partitionedRepos, err := PartitionRepos(request, s.NotSearcherOnly)
		if err != nil {
			return err
		}

		return runStructuralSearch(ctx, s.SearcherArgs, partitionedRepos, stream)
	})
}

func (*StructuralSearch) Name() string {
	return "Structural"
}

func (s *StructuralSearch) Required() bool {
	return s.IsRequired
}
