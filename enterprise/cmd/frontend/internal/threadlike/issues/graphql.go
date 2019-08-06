package issues

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/internal"
)

// GraphQLResolver implements the GraphQL Query and Mutation fields related to issues.
type GraphQLResolver struct{}

func init() {
	internal.ToGQLIssue = func(db *internal.DBThread) graphqlbackend.Issue { return newGQLIssue(db) }
}
