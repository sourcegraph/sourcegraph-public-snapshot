package commitstatuses

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func init() {
	graphqlbackend.CommitStatuses = GraphQLResolver{}
}
