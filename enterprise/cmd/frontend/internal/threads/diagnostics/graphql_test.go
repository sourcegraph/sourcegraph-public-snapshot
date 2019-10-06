package diagnostics

import "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"

func init() {
	graphqlbackend.ThreadDiagnostics = GraphQLResolver{}
}
