package structural

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	searchrepos "github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/searcher"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
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
	clients job.RuntimeClients
	repoSet repoData
	stream  streaming.Sender
}

// getJob returns a function parameterized by ctx to search over repos.
func (s *searchRepos) getJob(ctx context.Context) func() error {
	return func() error {
		searcherJob := &searcher.TextSearchJob{
			PatternInfo:     s.args.PatternInfo,
			Repos:           s.repoSet.AsList(),
			Indexed:         s.repoSet.IsIndexed(),
			UseFullDeadline: s.args.UseFullDeadline,
			Features:        s.args.Features,
		}

		_, err := searcherJob.Run(ctx, s.clients, s.stream)
		return err
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
func streamStructuralSearch(ctx context.Context, clients job.RuntimeClients, args *search.SearcherParameters, repos []repoData, stream streaming.Sender) (err error) {
	jobs := []*searchRepos{}
	for _, repoSet := range repos {
		searcherArgs := &search.SearcherParameters{
			PatternInfo:     args.PatternInfo,
			UseFullDeadline: args.UseFullDeadline,
			Features:        args.Features,
		}

		jobs = append(jobs, &searchRepos{clients: clients, args: searcherArgs, stream: stream, repoSet: repoSet})
	}
	return runJobs(ctx, jobs)
}

// retryStructuralSearch runs a structural search with a higher limit file match
// limit so that Zoekt resolves more potential file matches.
func retryStructuralSearch(ctx context.Context, clients job.RuntimeClients, args *search.SearcherParameters, repos []repoData, stream streaming.Sender) error {
	patternCopy := *(args.PatternInfo)
	patternCopy.FileMatchLimit = 1000
	argsCopy := *args
	argsCopy.PatternInfo = &patternCopy
	args = &argsCopy
	return streamStructuralSearch(ctx, clients, args, repos, stream)
}

func runStructuralSearch(ctx context.Context, clients job.RuntimeClients, args *search.SearcherParameters, batchRetry bool, repos []repoData, stream streaming.Sender) error {
	if !batchRetry {
		// stream search results
		return streamStructuralSearch(ctx, clients, args, repos, stream)
	}

	// For batching structural search we use retry logic if we get no results.
	agg := streaming.NewAggregatingStream()
	err := streamStructuralSearch(ctx, clients, args, repos, agg)

	event := agg.SearchEvent
	if len(event.Results) == 0 && err == nil {
		// retry structural search with a higher limit.
		aggRetry := streaming.NewAggregatingStream()
		err := retryStructuralSearch(ctx, clients, args, repos, aggRetry)
		if err != nil {
			// It is possible that the retry couldn't search any repos before the context
			// expired, in which case we send the stats from the first try.
			stats := aggRetry.Stats
			if stats.Zero() {
				stats = agg.Stats
			}
			stream.Send(streaming.SearchEvent{Stats: stats})
			return err
		}

		event = agg.SearchEvent
		if len(event.Results) == 0 {
			// Still no results? Give up.
			clients.Logger.Warn("Structural search gives up after more exhaustive attempt. Results may have been missed.")
			event.Stats.IsLimitHit = false // Ensure we don't display "Show more".
		}
	}

	matches := make([]result.Match, 0, len(event.Results))
	for _, fm := range event.Results {
		if _, ok := fm.(*result.FileMatch); !ok {
			return errors.Errorf("StructuralSearchJob failed to convert results")
		}
		matches = append(matches, fm)
	}

	stream.Send(streaming.SearchEvent{
		Results: matches,
		Stats:   event.Stats,
	})
	return err
}

type SearchJob struct {
	SearcherArgs     *search.SearcherParameters
	UseIndex         query.YesNoOnly
	ContainsRefGlobs bool
	BatchRetry       bool

	RepoOpts search.RepoOptions
}

func (s *SearchJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, stream, s)
	defer func() { finish(alert, err) }()

	repos := searchrepos.NewResolver(clients.Logger, clients.DB, clients.Gitserver, clients.SearcherURLs, clients.Zoekt)
	it := repos.Iterator(ctx, s.RepoOpts)

	for it.Next() {
		page := it.Current()
		page.MaybeSendStats(stream)

		indexed, unindexed, err := zoektutil.PartitionRepos(
			ctx,
			clients.Logger,
			page.RepoRevs,
			clients.Zoekt,
			search.TextRequest,
			s.UseIndex,
			s.ContainsRefGlobs,
		)
		if err != nil {
			return nil, err
		}

		repoSet := []repoData{UnindexedList(unindexed)}
		if indexed != nil {
			repoRevsFromBranchRepos := indexed.GetRepoRevsFromBranchRepos()
			repoSet = append(repoSet, IndexedMap(repoRevsFromBranchRepos))
		}
		err = runStructuralSearch(ctx, clients, s.SearcherArgs, s.BatchRetry, repoSet, stream)
		if err != nil {
			return nil, err
		}
	}

	return nil, it.Err()
}

func (*SearchJob) Name() string {
	return "StructuralSearchJob"
}

func (s *SearchJob) Attributes(v job.Verbosity) (res []attribute.KeyValue) {
	switch v {
	case job.VerbosityMax:
		res = append(res,
			attribute.Bool("useFullDeadline", s.SearcherArgs.UseFullDeadline),
			attribute.Bool("containsRefGlobs", s.ContainsRefGlobs),
			attribute.String("useIndex", string(s.UseIndex)),
		)
		fallthrough
	case job.VerbosityBasic:
		res = append(res, trace.Scoped("patternInfo", s.SearcherArgs.PatternInfo.Fields()...)...)
		res = append(res, trace.Scoped("repoOpts", s.RepoOpts.Attributes()...)...)
	}
	return res
}

func (s *SearchJob) Children() []job.Describer       { return nil }
func (s *SearchJob) MapChildren(job.MapFunc) job.Job { return s }
