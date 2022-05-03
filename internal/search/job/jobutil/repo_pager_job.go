package jobutil

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
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
	child            job.Job         // child job tree that need populating a repos field to run
}

// setRepos populates the repos field for all jobs that need repos. Jobs are
// copied, ensuring this function is side-effect free.
func setRepos(job job.Job, indexed *zoekt.IndexedRepoRevs, unindexed []*search.RepositoryRevisions) job.Job {
	setZoektRepos := func(job *zoekt.ZoektRepoSubsetSearchJob) *zoekt.ZoektRepoSubsetSearchJob {
		jobCopy := *job
		jobCopy.Repos = indexed
		return &jobCopy
	}

	setSearcherRepos := func(job *searcher.SearcherJob) *searcher.SearcherJob {
		jobCopy := *job
		jobCopy.Repos = unindexed
		return &jobCopy
	}

	setZoektSymbolRepos := func(job *zoekt.ZoektSymbolSearchJob) *zoekt.ZoektSymbolSearchJob {
		jobCopy := *job
		jobCopy.Repos = indexed
		return &jobCopy
	}

	setSymbolSearcherRepos := func(job *searcher.SymbolSearcherJob) *searcher.SymbolSearcherJob {
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

func (p *repoPagerJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, stream, p)
	defer func() { finish(alert, err) }()

	var maxAlerter search.MaxAlerter

	repoResolver := &repos.Resolver{DB: clients.DB, Opts: p.repoOptions}
	pager := func(page *repos.Resolved) error {
		indexed, unindexed, err := zoekt.PartitionRepos(
			ctx,
			page.RepoRevs,
			clients.Zoekt,
			search.TextRequest,
			p.useIndex,
			p.containsRefGlobs,
		)
		if err != nil {
			return err
		}

		job := setRepos(p.child, indexed, unindexed)
		alert, err := job.Run(ctx, clients, stream)
		maxAlerter.Add(alert)
		return err
	}

	return maxAlerter.Alert, repoResolver.Paginate(ctx, pager)
}

func (p *repoPagerJob) Name() string {
	return "RepoPager"
}
