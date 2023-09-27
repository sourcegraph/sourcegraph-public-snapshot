package jobutil

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/iterator"
)

// Exhaustive exports what is needed for the search jobs product (exhaustive
// search). The naming conflict between the product search jobs and the search
// job infrastructure is unfortunate. So we use the name exhaustive to
// differentiate ourself from the infrastructure.
type Exhaustive struct {
	repoPagerJob *repoPagerJob
}

// NewExhaustive constructs Exhaustive from the search inputs.
//
// It will return an error if the input query is not supported by Exhaustive.
func NewExhaustive(inputs *search.Inputs) (Exhaustive, error) {
	// TODO(keegan) a bunch of tests around this after branch cut pls

	if inputs.Protocol != search.Exhaustive {
		return Exhaustive{}, errors.New("only works for exhaustive search inputs")
	}

	if len(inputs.Plan) != 1 {
		return Exhaustive{}, errors.Errorf("expected a simple expression (no and/or/etc). Got multiple jobs to run %v", inputs.Plan)
	}

	b := inputs.Plan[0]
	term, ok := b.Pattern.(query.Pattern)
	if !ok {
		return Exhaustive{}, errors.Errorf("expected a simple expression (no and/or/etc). Got %v", b.Pattern)
	}

	planJob, err := NewFlatJob(inputs, query.Flat{Parameters: b.Parameters, Pattern: &term})
	if err != nil {
		return Exhaustive{}, err
	}

	repoPagerJob, ok := planJob.(*repoPagerJob)
	if !ok {
		return Exhaustive{}, errors.Errorf("internal error: expected a repo pager job when converting plan into search jobs got %T", planJob)
	}

	return Exhaustive{
		repoPagerJob: repoPagerJob,
	}, nil
}

func (e Exhaustive) Job(repoRevs *search.RepositoryRevisions) job.Job {
	// TODO should we add in a timeout and limit here?
	// TODO should we support indexed search and run through zoekt.PartitionRepos?
	return e.repoPagerJob.child.Resolve(resolvedRepos{
		unindexed: []*search.RepositoryRevisions{repoRevs},
	})
}

// RepositoryRevSpecs is a wrapper around repos.Resolver.IterateRepoRevs.
func (e Exhaustive) RepositoryRevSpecs(ctx context.Context, clients job.RuntimeClients) *iterator.Iterator[repos.RepoRevSpecs] {
	return reposNewResolver(clients).IterateRepoRevs(ctx, e.repoPagerJob.repoOpts)
}

// ResolveRepositoryRevSpec is a wrapper around repos.Resolver.ResolveRevSpecs.
func (e Exhaustive) ResolveRepositoryRevSpec(ctx context.Context, clients job.RuntimeClients, repoRevSpecs []repos.RepoRevSpecs) (repos.Resolved, error) {
	return reposNewResolver(clients).ResolveRevSpecs(ctx, e.repoPagerJob.repoOpts, repoRevSpecs)
}

func reposNewResolver(clients job.RuntimeClients) *repos.Resolver {
	return repos.NewResolver(clients.Logger, clients.DB, clients.Gitserver, clients.SearcherURLs, clients.Zoekt)
}
