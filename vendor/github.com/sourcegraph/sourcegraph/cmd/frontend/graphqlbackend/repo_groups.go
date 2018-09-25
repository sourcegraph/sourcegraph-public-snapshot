package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

type repoGroup struct {
	name         string
	repositories []api.RepoURI
}

func (g repoGroup) Name() string { return g.name }

func (g repoGroup) Repositories() []string { return repoURIsToStrings(g.repositories) }

func (r *schemaResolver) RepoGroups(ctx context.Context) ([]*repoGroup, error) {
	groupsByName, err := resolveRepoGroups(ctx)
	if err != nil {
		return nil, err
	}

	groups := make([]*repoGroup, 0, len(groupsByName))
	for name, repos := range groupsByName {
		repoPaths := make([]api.RepoURI, len(repos))
		for i, repo := range repos {
			repoPaths[i] = repo.URI
		}
		groups = append(groups, &repoGroup{
			name:         name,
			repositories: repoPaths,
		})
	}
	return groups, nil
}
