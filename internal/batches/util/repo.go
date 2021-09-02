package util

import (
	"github.com/sourcegraph/sourcegraph/lib/batches/template"

	"github.com/sourcegraph/src-cli/internal/batches/graphql"
)

// GraphQLRepoToTemplatingRepo transforms a given *graphql.Repository into a
// template.Repository.
func GraphQLRepoToTemplatingRepo(r *graphql.Repository) template.Repository {
	return template.Repository{
		Name:        r.Name,
		FileMatches: r.FileMatches,
	}
}
