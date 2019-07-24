package campaigns

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func init() {
	graphqlbackend.Campaigns = GraphQLResolver{}
}
