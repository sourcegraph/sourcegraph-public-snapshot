package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

type repoGroup struct {
	name         string
	repositories []api.RepoName
}

func (g repoGroup) Name() string { return g.name }

func (g repoGroup) Repositories() []string { return repoNamesToStrings(g.repositories) }

func (r *schemaResolver) RepoGroups(ctx context.Context) ([]*repoGroup, error) {
	settings, err := decodedViewerFinalSettings(ctx)
	if err != nil {
		return nil, err
	}

	groupsByName, err, _ := resolveRepoGroups(settings)
	if err != nil {
		return nil, err
	}

	groups := make([]*repoGroup, 0, len(groupsByName))
	for name, repos := range groupsByName {
		repoPaths := make([]api.RepoName, len(repos))
		for i, repo := range repos {
			repoPaths[i] = repo.Name
		}
		groups = append(groups, &repoGroup{
			name:         name,
			repositories: repoPaths,
		})
	}
	return groups, nil
}
