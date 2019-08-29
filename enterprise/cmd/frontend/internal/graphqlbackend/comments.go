package graphqlbackend

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
)

func init() {
	// Contribute the GraphQL types for comments.
	graphqlbackend.Comments = comments.GraphQLResolver{}
}
