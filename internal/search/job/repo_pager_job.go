package job

import (
	"context"

	zoektstreamer "github.com/google/zoekt"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/searcher"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/search/zoekt"
)

type repoPagerJob struct {
	repoOptions      search.RepoOptions
	useIndex         query.YesNoOnly // whether to include indexed repos
	containsRefGlobs bool            // whether to include repositories with refs
	child            Job             // child job tree that need populating a repos field to run

	zoekt zoektstreamer.Streamer
}

// setRepos populates the repos field for all jobs that need repos. Jobs are
// copied, ensuring this function is side-effect free.
func setRepos(job Job, indexed *zoekt.IndexedRepoRevs, unindexed []*search.RepositoryRevisions) Job {
	setZoektRepos := func(job *zoekt.ZoektRepoSubsetSearch) *zoekt.ZoektRepoSubsetSearch {
		jobCopy := *job
		jobCopy.Repos = indexed
		return &jobCopy
	}

	setSearcherRepos := func(job *searcher.Searcher) *searcher.Searcher {
		jobCopy := *job
		jobCopy.Repos = unindexed
		return &jobCopy
	}

	setZoektSymbolRepos := func(job *zoekt.ZoektSymbolSearch) *zoekt.ZoektSymbolSearch {
		jobCopy := *job
		jobCopy.Repos = indexed
		return &jobCopy
	}

	setSymbolSearcherRepos := func(job *searcher.SymbolSearcher) *searcher.SymbolSearcher {
		jobCopy := *job
		jobCopy.Repos = unindexed
		return &jobCopy
	}

	setRepos := Mapper{
		MapZoektRepoSubsetSearchJob: setZoektRepos,
		MapZoektSymbolSearchJob:     setZoektSymbolRepos,
		MapSearcherJob:              setSearcherRepos,
		MapSymbolSearcherJob:        setSymbolSearcherRepos,
	}

	return setRepos.Map(job)
}

func (p *repoPagerJob) Run(ctx context.Context, db database.DB, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := jobutil.StartSpan(ctx, stream, p)
	defer func() { finish(alert, err) }()

	var maxAlerter search.MaxAlerter

	repoResolver := &repos.Resolver{DB: db, Opts: p.repoOptions}
	pager := func(page *repos.Resolved) error {
		indexed, unindexed, err := zoekt.PartitionRepos(
			ctx,
			page.RepoRevs,
			p.zoekt,
			search.TextRequest,
			p.useIndex,
			p.containsRefGlobs,
		)
		if err != nil {
			return err
		}

		job := setRepos(p.child, indexed, unindexed)
		alert, err := job.Run(ctx, db, stream)
		maxAlerter.Add(alert)
		return err
	}

	return maxAlerter.Alert, repoResolver.Paginate(ctx, nil, pager)
}

func (p *repoPagerJob) Name() string {
	return "RepoPager"
}
