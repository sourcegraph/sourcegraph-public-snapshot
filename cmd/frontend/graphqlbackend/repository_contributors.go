package graphqlbackend

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type repositoryContributorsArgs struct {
	RevisionRange *string
	After         *string
	Path          *string
}

func (r *RepositoryResolver) Contributors(args *struct {
	repositoryContributorsArgs
	First *int32
}) *repositoryContributorConnectionResolver {
	return &repositoryContributorConnectionResolver{
		db:    r.db,
		args:  args.repositoryContributorsArgs,
		first: args.First,
		repo:  r,
	}
}

type repositoryContributorConnectionResolver struct {
	db    dbutil.DB
	args  repositoryContributorsArgs
	first *int32

	repo *RepositoryResolver

	// cache result because it is used by multiple fields
	once    sync.Once
	results []*git.PersonCount
	err     error
}

func (r *repositoryContributorConnectionResolver) compute(ctx context.Context) ([]*git.PersonCount, error) {
	r.once.Do(func() {
		var opt git.ShortLogOptions
		if r.args.RevisionRange != nil {
			opt.Range = *r.args.RevisionRange
		}
		if r.args.Path != nil {
			opt.Path = *r.args.Path
		}
		if r.args.After != nil {
			opt.After = *r.args.After
		}
		r.results, r.err = git.ShortLog(ctx, r.repo.name, opt)
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
			db:    r.db,
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

func (r *repositoryContributorConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	results, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(r.first != nil && len(results) > int(*r.first)), nil
}
