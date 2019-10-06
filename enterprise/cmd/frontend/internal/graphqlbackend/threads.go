package graphqlbackend

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads/diagnostics"
)

func init() {
	// Contribute the GraphQL types for threads.
	graphqlbackend.Threads = threads.GraphQLResolver{}
	graphqlbackend.ThreadDiagnostics = diagnostics.GraphQLResolver{}
}
