package events

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func init() {
	graphqlbackend.Events = GraphQLResolver{}
}
