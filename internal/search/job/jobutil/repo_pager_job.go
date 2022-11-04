package jobutil

import (
	"context"

	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/searcher"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/search/zoekt"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type repoPagerJob struct {
	repoOpts         search.RepoOptions
	containsRefGlobs bool                          // whether to include repositories with refs
	child            job.PartialJob[resolvedRepos] // child job tree that need populating a repos field to run
}

// resolvedRepos is the set of information to complete the partial
// child jobs for the repoPagerJob.
type resolvedRepos struct {
	indexed   *zoekt.IndexedRepoRevs
	unindexed []*search.RepositoryRevisions
}

// reposPartialJob is a partial job that needs a set of resolved repos
// in order to construct a complete job.
type reposPartialJob struct {
	inner job.Job
}

func (j *reposPartialJob) Partial() job.Job {
	return j.inner
}

func (j *reposPartialJob) Resolve(rr resolvedRepos) job.Job {
	return setRepos(j.inner, rr.indexed, rr.unindexed)
}

func (j *reposPartialJob) Name() string                       { return "PartialReposJob" }
func (j *reposPartialJob) Fields(job.Verbosity) []otlog.Field { return nil }
func (j *reposPartialJob) Children() []job.Describer          { return []job.Describer{j.inner} }
func (j *reposPartialJob) MapChildren(fn job.MapFunc) job.PartialJob[resolvedRepos] {
	cp := *j
	cp.inner = job.Map(j.inner, fn)
	return &cp
}

// setRepos populates the repos field for all jobs that need repos. Jobs are
// copied, ensuring this function is side-effect free.
func setRepos(j job.Job, indexed *zoekt.IndexedRepoRevs, unindexed []*search.RepositoryRevisions) job.Job {
	return job.Map(j, func(j job.Job) job.Job {
		switch v := j.(type) {
		case *zoekt.RepoSubsetTextSearchJob:
			cp := *v
			cp.Repos = indexed
			return &cp
		case *searcher.TextSearchJob:
			cp := *v
			cp.Repos = unindexed
			return &cp
		case *zoekt.SymbolSearchJob:
			cp := *v
			cp.Repos = indexed
			return &cp
		case *searcher.SymbolSearchJob:
			cp := *v
			cp.Repos = unindexed
			return &cp
		default:
			return j
		}
	})
}

func (p *repoPagerJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, stream, p)
	defer func() { finish(alert, err) }()

	var maxAlerter search.MaxAlerter

	repoResolver := repos.NewResolver(clients.Logger, clients.DB, clients.Gitserver, clients.SearcherURLs, clients.Zoekt)
	pager := func(page *repos.Resolved) error {
		indexed, unindexed, err := zoekt.PartitionRepos(
			ctx,
			clients.Logger,
			page.RepoRevs,
			clients.Zoekt,
			search.TextRequest,
			p.repoOpts.UseIndex,
			p.containsRefGlobs,
		)
		if err != nil {
			return err
		}

		job := p.child.Resolve(resolvedRepos{indexed, unindexed})
		alert, err := job.Run(ctx, clients, stream)
		maxAlerter.Add(alert)
		return err
	}

	return maxAlerter.Alert, repoResolver.Paginate(ctx, p.repoOpts, pager)
}

func (p *repoPagerJob) Name() string {
	return "RepoPagerJob"
}

func (p *repoPagerJob) Fields(v job.Verbosity) (res []otlog.Field) {
	switch v {
	case job.VerbosityMax:
		res = append(res,
			otlog.Bool("containsRefGlobs", p.containsRefGlobs),
		)
		fallthrough
	case job.VerbosityBasic:
		res = append(res,
			trace.Scoped("repoOpts", p.repoOpts.Tags()...),
		)
	}
	return res
}

func (p *repoPagerJob) Children() []job.Describer {
	return []job.Describer{p.child}
}

func (p *repoPagerJob) MapChildren(fn job.MapFunc) job.Job {
	cp := *p
	cp.child = p.child.MapChildren(fn)
	return &cp
}
