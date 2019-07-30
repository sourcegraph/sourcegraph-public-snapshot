package graphqlbackend

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/changesets"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/threads"
)

func init() {
	// Contribute the GraphQL types for threads, issues, and changesets.
	graphqlbackend.ThreadOrIssueOrChangesets = threadlike.GraphQLResolver{}
	graphqlbackend.Threads = threads.GraphQLResolver{}
	graphqlbackend.Changesets = changesets.GraphQLResolver{}
}
