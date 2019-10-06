package comments

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func init() {
	graphqlbackend.Comments = GraphQLResolver{}
}
