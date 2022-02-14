package structural

import (
	"context"

	"github.com/inconshreveable/log15"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/limits"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	searchrepos "github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/search/textsearch"
	zoektutil "github.com/sourcegraph/sourcegraph/internal/search/zoekt"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
		return textsearch.CallSearcherOverRepos(ctx, s.args, s.stream, s.repoSet.AsList(), s.repoSet.IsIndexed())
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
	if args.PatternInfo.FileMatchLimit != limits.DefaultMaxSearchResults {
		// streamStructuralSearch performs a streaming search when the user sets a value
		// for `count`. The first return parameter indicates whether the request was
		// serviced with streaming.
		return streamStructuralSearch(ctx, args, repos, stream)
	}

	// For structural search with default limits we retry if we get no results.
	agg := streaming.NewAggregatingStream()
	err := streamStructuralSearch(ctx, args, repos, agg)

	event := agg.SearchEvent
	if len(event.Results) == 0 && err == nil {
		// retry structural search with a higher limit.
		agg := streaming.NewAggregatingStream()
		err := retryStructuralSearch(ctx, args, repos, agg)
		if err != nil {
			return err
		}

		event = agg.SearchEvent
		if len(event.Results) == 0 {
			// Still no results? Give up.
			log15.Warn("Structural search gives up after more exhaustive attempt. Results may have been missed.")
			event.Stats.IsLimitHit = false // Ensure we don't display "Show more".
		}
	}

	matches := make([]result.Match, 0, len(event.Results))
	for _, fm := range event.Results {
		if _, ok := fm.(*result.FileMatch); !ok {
			return errors.Errorf("StructuralSearch failed to convert results")
		}
		matches = append(matches, fm)
	}

	stream.Send(streaming.SearchEvent{
		Results: matches,
		Stats:   event.Stats,
	})
	return err
}

type StructuralSearch struct {
	ZoektArgs    *search.ZoektParameters
	SearcherArgs *search.SearcherParameters

	NotSearcherOnly  bool
	UseIndex         query.YesNoOnly
	ContainsRefGlobs bool

	RepoOpts search.RepoOptions
}

func (s *StructuralSearch) Run(ctx context.Context, db database.DB, stream streaming.Sender) (_ *search.Alert, err error) {
	tr, ctx := trace.New(ctx, "StructuralSearch", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	repos := &searchrepos.Resolver{DB: db, Opts: s.RepoOpts}
	return nil, repos.Paginate(ctx, nil, func(page *searchrepos.Resolved) error {
		request, ok, err := zoektutil.OnlyUnindexed(page.RepoRevs, s.ZoektArgs.Zoekt, s.UseIndex, s.ContainsRefGlobs, zoektutil.MissingRepoRevStatus(stream))
		if err != nil {
			return err
		}
		if !ok {
			request, err = zoektutil.NewIndexedSubsetSearchRequest(ctx, page.RepoRevs, s.UseIndex, s.ZoektArgs, zoektutil.MissingRepoRevStatus(stream))
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
