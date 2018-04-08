package graphqlbackend

import (
	"context"
	"sync"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type repositoryContributorsArgs struct {
	Range *string
	Path  *string
}

func (r *repositoryResolver) Contributors(args *struct {
	repositoryContributorsArgs
	First *int32
}) *repositoryContributorConnectionResolver {
	return &repositoryContributorConnectionResolver{
		args:  args.repositoryContributorsArgs,
		first: args.First,
		repo:  r,
	}
}

type repositoryContributorConnectionResolver struct {
	args  repositoryContributorsArgs
	first *int32

	repo *repositoryResolver

	// cache result because it is used by multiple fields
	once    sync.Once
	results []*vcs.PersonCount
	err     error
}

func (r *repositoryContributorConnectionResolver) compute(ctx context.Context) ([]*vcs.PersonCount, error) {
	r.once.Do(func() {
		var opt vcs.ShortLogOptions
		if r.args.Range != nil {
			opt.Range = *r.args.Range
		}

		vcsrepo := backend.Repos.CachedVCS(r.repo.repo)
		r.results, r.err = vcsrepo.ShortLog(ctx, opt)
	})
	return r.results, r.err
}

func (r *repositoryContributorConnectionResolver) Nodes(ctx context.Context) ([]*repositoryContributorResolver, error) {
	results, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if r.first != nil && len(results) > int(*r.first) {
		results = results[:*r.first]
	}

	resolvers := make([]*repositoryContributorResolver, len(results))
	for i, contributor := range results {
		resolvers[i] = &repositoryContributorResolver{
			name:  contributor.Name,
			email: contributor.Email,
			count: contributor.Count,
			repo:  r.repo,
			args:  r.args,
		}
	}
	return resolvers, nil
}

func (r *repositoryContributorConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	results, err := r.compute(ctx)
	if err != nil {
		return 0, err
	}
	return int32(len(results)), nil
}

func (r *repositoryContributorConnectionResolver) PageInfo(ctx context.Context) (*pageInfo, error) {
	results, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return &pageInfo{hasNextPage: r.first != nil && len(results) > int(*r.first)}, nil
}
