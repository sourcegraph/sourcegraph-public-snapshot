package git

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

// GraphQLResolver implements the GraphQL Query and Mutation fields related to git.
type GraphQLResolver struct{}

func init() {
	// Contribute the GraphQL type GitMutation.
	graphqlbackend.Git = GraphQLResolver{}
}
