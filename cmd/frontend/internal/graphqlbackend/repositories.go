package graphqlbackend

import (
	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
)

func (r *schemaResolver) Repositories(ctx context.Context) (*repositoryConnectionResolver, error) {
	opt := &sourcegraph.RepoListOptions{}
	opt.PerPage = 10000 // we want every repo
	repos, err := listRepos(ctx, opt)
	if err != nil {
		return nil, err
	}
	return &repositoryConnectionResolver{repos: repos}, nil
}

func listRepos(ctx context.Context, opt *sourcegraph.RepoListOptions) ([]*repositoryResolver, error) {
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

type repositoryConnectionResolver struct {
	repos []*repositoryResolver
}

func (r *repositoryConnectionResolver) Nodes() []*repositoryResolver { return r.repos }

func (r *repositoryConnectionResolver) TotalCount() int32 { return int32(len(r.repos)) }
