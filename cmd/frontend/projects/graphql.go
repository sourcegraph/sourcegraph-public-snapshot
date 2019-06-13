package projects

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

// GraphQLResolver implements the GraphQL Query and Mutation fields related to projects.
type GraphQLResolver struct{}

func init() {
	// Contribute the GraphQL type ProjectsMutation.
	graphqlbackend.Projects = GraphQLResolver{}
}
