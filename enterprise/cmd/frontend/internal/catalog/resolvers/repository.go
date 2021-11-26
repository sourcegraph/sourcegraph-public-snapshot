package resolvers

import (
	"context"
	"path"
	"strings"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/catalog"
)

func (r *rootResolver) RepositoryComponents(ctx context.Context, repo *gql.RepositoryResolver, args *gql.RepositoryComponentsArgs) ([]gql.ComponentResolver, error) {
	var matches []gql.ComponentResolver
	components := catalog.Components()
componentLoop:
	for _, c := range components {
		for _, sloc := range c.SourceLocations {
			if sloc.Repo == repo.RepoName() {
				for _, p := range sloc.Paths {
					var match bool
					if args.Recursive {
						pathContainsComponent := pathHasPrefix(p, args.Path)
						match = pathContainsComponent
					} else {
						match = path.Clean(args.Path) == path.Clean(p)
					}
					if match {
						matches = append(matches, &componentResolver{db: r.db, component: c})
						continue componentLoop // only add each one component once
					}
				}
			}
		}
	}
	return matches, nil
}

func (r *rootResolver) GitTreeEntryComponents(ctx context.Context, treeEntry *gql.GitTreeEntryResolver, args *gql.GitTreeEntryComponentsArgs) ([]gql.ComponentResolver, error) {
	var matches []gql.ComponentResolver
	components := catalog.Components()
	for _, c := range components {
		for _, sloc := range c.SourceLocations {
			if sloc.Repo == treeEntry.Repository().RepoName() {
				for _, p := range sloc.Paths {
					if componentContainsTreeEntry := pathHasPrefix(treeEntry.Path(), p); componentContainsTreeEntry {
						matches = append(matches, &componentResolver{db: r.db, component: c})
					}
				}
			}
		}
	}
	return matches, nil
}

func pathHasPrefix(p, prefix string) bool {
	prefix = path.Clean(prefix)
	if prefix == "." {
		return true
	}
	return p == prefix || strings.HasPrefix(p, prefix+"/")
}
