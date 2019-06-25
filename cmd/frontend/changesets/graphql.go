package changesets

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

// GraphQLResolver implements the GraphQL Query and Mutation fields related to changesets.
type GraphQLResolver struct{}

func init() {
	// Contribute the GraphQL type ChangesetsMutation.
	graphqlbackend.Changesets = GraphQLResolver{}
}
