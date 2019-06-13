package graphqlbackend

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/rules"
)

func init() {
	// Contribute the GraphQL type RulesMutation.
	graphqlbackend.Rules = rules.GraphQLResolver{}
}
