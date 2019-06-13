package labels

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"

	// This package's tests rely on GraphQL resolvers registered as a side effect of importing
	// package projects.
	_ "github.com/sourcegraph/sourcegraph/cmd/frontend/projects"
)

func init() {
	graphqlbackend.Labels = GraphQLResolver{}
}
