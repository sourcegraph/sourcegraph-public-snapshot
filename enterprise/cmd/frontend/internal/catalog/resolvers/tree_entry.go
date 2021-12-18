package resolvers

import (
	"context"
	"strings"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/catalog"
)

func (r *rootResolver) GitTreeEntryComponents(ctx context.Context, treeEntry *gql.GitTreeEntryResolver) ([]gql.ComponentResolver, error) {
	var matches []gql.ComponentResolver

	entities, _, _ := catalog.Data()
	for _, e := range entities {
		// TODO(sqs): dont require match on commit
		if e.SourceRepo == treeEntry.Repository().RepoName() {
			for _, p := range e.SourcePaths {
				if p == treeEntry.Path() || strings.HasPrefix(treeEntry.Path(), p+"/") {
					matches = append(matches, &componentResolver{db: r.db, component: e})
				}
			}
		}
	}

	return matches, nil
}
