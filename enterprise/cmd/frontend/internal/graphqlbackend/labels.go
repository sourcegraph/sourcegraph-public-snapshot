package graphqlbackend

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/labels"
)

func init() {
	// Contribute the GraphQL type LabelsMutation.
	graphqlbackend.Labels = labels.GraphQLResolver{}
}
