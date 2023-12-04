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
	if inputs.Protocol != search.Exhaustive {
		return Exhaustive{}, errors.New("only works for exhaustive search inputs")
	}

	// This doesn't lead to an error, but we will drop result types other than
	// "file" which might be surprising to users.
	types, _ := inputs.Query.StringValues(query.FieldType)
	if len(types) != 1 || types[0] != "file" {
		return Exhaustive{}, errors.Errorf("expected \"type:file\" only. Got %v", types)
	}

	if len(inputs.Plan) != 1 {
		return Exhaustive{}, errors.Errorf("expected a simple expression (no and/or/etc). Got multiple jobs to run %v", inputs.Plan)
	}

	b := inputs.Plan[0]
	term, ok := b.Pattern.(query.Pattern)
	if !ok {
		return Exhaustive{}, errors.Errorf("expected a simple expression (no and/or/etc). Got %v", b.Pattern)
	}

	// We don't support file: predicates, such as file:has.content(), because the
	// search breaks in unexpected ways. For example, for interactive search
	// file:has.content() is translated to an AND query which we don't support in
	// Search Jobs yet.
	if pred, ok := hasPredicates(query.FieldFile, inputs.Query); ok {
		return Exhaustive{}, errors.Errorf("file: predicates are not supported. Got %v", pred)
	}

	// This is a very weak protection but should be enough to catch simple misuse.
	if inputs.PatternType == query.SearchTypeRegex && term.Value == ".*" {
		return Exhaustive{}, errors.Errorf("regex search with .* is not supported")
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

func hasPredicates(field string, q query.Q) (pred string, ok bool) {
	values, negated := q.StringValues(field)
	for _, v := range append(values, negated...) {
		pred, _, ok = query.ScanPredicate(field, []byte(v), query.DefaultPredicateRegistry)
		if ok {
			break
		}
	}
	return
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
