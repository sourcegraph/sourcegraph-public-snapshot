package threads

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/internal"
)

// GraphQLResolver implements the GraphQL Query and Mutation fields related to threads.
type GraphQLResolver struct{}

func init() {
	internal.ToGQLThread = func(db *internal.DBThread) graphqlbackend.Thread { return newGQLThread(db) }
}
