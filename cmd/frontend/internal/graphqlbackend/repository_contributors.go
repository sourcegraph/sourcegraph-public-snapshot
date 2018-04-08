package graphqlbackend

import (
	"context"
	"sync"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

func (r *repositoryResolver) Contributors(args *struct {
	Range *string
	First *int32
}) *contributorConnectionResolver {
	return &contributorConnectionResolver{
		range_: args.Range,
		first:  args.First,
		repo:   r,
	}
}

type contributorConnectionResolver struct {
	range_ *string
	first  *int32

	repo *repositoryResolver

	// cache result because it is used by multiple fields
	once    sync.Once
	results []*vcs.PersonCount
	err     error
}

func (r *contributorConnectionResolver) compute(ctx context.Context) ([]*vcs.PersonCount, error) {
	r.once.Do(func() {
		var opt vcs.ShortLogOptions
		if r.range_ != nil {
			opt.Range = *r.range_
		}

		vcsrepo := backend.Repos.CachedVCS(r.repo.repo)
		r.results, r.err = vcsrepo.ShortLog(ctx, opt)
	})
	return r.results, r.err
}

func (r *contributorConnectionResolver) Nodes(ctx context.Context) ([]*contributorResolver, error) {
	results, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if r.first != nil && len(results) > int(*r.first) {
		results = results[:*r.first]
	}

	resolvers := make([]*contributorResolver, len(results))
	for i, contributor := range results {
		resolvers[i] = &contributorResolver{
			name:  contributor.Name,
			email: contributor.Email,
			count: contributor.Count,
		}
	}
	return resolvers, nil
}

func (r *contributorConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	results, err := r.compute(ctx)
	if err != nil {
		return 0, err
	}
	return int32(len(results)), nil
}

func (r *contributorConnectionResolver) PageInfo(ctx context.Context) (*pageInfo, error) {
	results, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return &pageInfo{hasNextPage: r.first != nil && len(results) > int(*r.first)}, nil
}
