package changesets

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/internal"
)

// GraphQLResolver implements the GraphQL Query and Mutation fields related to changesets.
type GraphQLResolver struct{}

func init() {
	internal.ToGQLChangeset = func(db *internal.DBThread) graphqlbackend.Changeset { return newGQLChangeset(db) }
}
