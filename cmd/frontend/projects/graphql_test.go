package projects

import "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"

func init() {
	graphqlbackend.Projects = GraphQLResolver{}
}
