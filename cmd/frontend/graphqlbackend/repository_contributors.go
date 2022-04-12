package graphqlbackend

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
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
	db    database.DB
	args  repositoryContributorsArgs
	first *int32

	repo *RepositoryResolver

	// cache result because it is used by multiple fields
	once    sync.Once
	results []*gitdomain.PersonCount
	err     error
}

func (r *repositoryContributorConnectionResolver) compute(ctx context.Context) ([]*gitdomain.PersonCount, error) {
	r.once.Do(func() {
		client := gitserver.NewClient(r.db)
		var opt gitserver.ShortLogOptions
		if r.args.RevisionRange != nil {
			opt.Range = *r.args.RevisionRange
		}
		if r.args.Path != nil {
			opt.Path = *r.args.Path
		}
		if r.args.After != nil {
			opt.After = *r.args.After
		}
		r.results, r.err = client.ShortLog(ctx, r.repo.RepoName(), opt)
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
