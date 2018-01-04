package graphqlbackend

import (
	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
)

func (r *siteResolver) Repositories(args *struct {
	connectionArgs
	Query *string
}) *repositoryConnectionResolver {
	return &repositoryConnectionResolver{
		connectionResolverCommon: newConnectionResolverCommon(args.connectionArgs),
		query: args.Query,
	}
}

type repositoryConnectionResolver struct {
	connectionResolverCommon
	query *string
}

func (r *repositoryConnectionResolver) Nodes(ctx context.Context) ([]*repositoryResolver, error) {
	opt := &sourcegraph.RepoListOptions{
		ListOptions: sourcegraph.ListOptions{
			PerPage: r.first,
		},
	}
	if r.query != nil {
		opt.Query = *r.query
	}

	reposList, err := backend.Repos.List(ctx, opt)
	if err != nil {
		return nil, err
	}

	var l []*repositoryResolver
	for _, repo := range reposList.Repos {
		l = append(l, &repositoryResolver{
			repo: repo,
		})
	}
	return l, nil
}

func (r *repositoryConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	count, err := db.Repos.Count(ctx)
	return int32(count), err
}
