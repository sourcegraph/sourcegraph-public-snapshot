package graphql

import (
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

type IndexConfigurationResolver struct {
	configuration []byte
}

func NewIndexConfigurationResolver(configuration []byte) gql.IndexConfigurationResolver {
	return &IndexConfigurationResolver{
		configuration: configuration,
	}
}

func (r *IndexConfigurationResolver) Configuration() *string {
	return strPtr(string(r.configuration))
}
