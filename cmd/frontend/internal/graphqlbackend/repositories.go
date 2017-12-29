package graphqlbackend

import (
	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

func (r *schemaResolver) Repositories(args *struct {
	First *int32
}) *repositoryConnectionResolver {
	var c repositoryConnectionResolver

	if args.First == nil {
		c.first = defaultFirstValue
	} else {
		c.first = *args.First
	}

	return &c
}

type repositoryConnectionResolver struct {
	first int32
}

func (r *repositoryConnectionResolver) Nodes(ctx context.Context) ([]*repositoryResolver, error) {
	reposList, err := backend.Repos.List(ctx, &sourcegraph.RepoListOptions{
		ListOptions: sourcegraph.ListOptions{
			PerPage: r.first,
		},
	})
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
	count, err := localstore.Repos.Count(ctx)
	return int32(count), err
}
