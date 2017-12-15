package graphqlbackend

import "context"

type repoGroup struct {
	name         string
	repositories []string
}

func (g repoGroup) Name() string { return g.name }

func (g repoGroup) Repositories() []string { return g.repositories }

func (r *schemaResolver) RepoGroups(ctx context.Context) ([]*repoGroup, error) {
	groupsByName, err := resolveRepoGroups(ctx)
	if err != nil {
		return nil, err
	}

	groups := make([]*repoGroup, 0, len(groupsByName))
	for name, repos := range groupsByName {
		repoPaths := make([]string, len(repos))
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
