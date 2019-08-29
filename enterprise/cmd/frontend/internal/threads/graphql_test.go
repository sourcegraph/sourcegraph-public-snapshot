package threads

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func init() {
	graphqlbackend.Threads = GraphQLResolver{}
}
