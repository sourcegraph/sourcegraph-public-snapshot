package resolvers

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/catalog"
)

func (r *componentResolver) DescendentComponents(ctx context.Context) ([]gql.ComponentResolver, error) {
	slocs, err := r.sourceSetResolver(ctx)
	if err != nil {
		return nil, err
	}
	return slocs.DescendentComponents(ctx)
}

func (r *rootResolver) GitTreeEntryDescendentComponents(ctx context.Context, treeEntry *gql.GitTreeEntryResolver) ([]gql.ComponentResolver, error) {
	return sourceSetResolverFromTreeEntry(treeEntry, r.db).DescendentComponents(ctx)
}

func (r *sourceSetResolver) DescendentComponents(ctx context.Context) ([]gql.ComponentResolver, error) {
	var matches []gql.ComponentResolver
	components := catalog.Components()
	for _, c := range components {
		for _, cSloc := range c.SourceLocations {
			for _, rSloc := range r.slocs {
				if cSloc.Repo == rSloc.repoName {
					for _, p := range cSloc.Paths {
						if p == rSloc.path {
							continue
						}
						if pathContainsComponent := pathHasPrefix(p, rSloc.path); pathContainsComponent {
							matches = append(matches, &componentResolver{db: r.db, component: c})
						}
					}
				}
			}
		}
	}
	return matches, nil
}
