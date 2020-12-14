package graphql

import (
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
)

type IndexConfigurationResolver struct {
	configuration store.IndexConfiguration
}

func NewIndexConfigurationResolver(configuration store.IndexConfiguration) gql.IndexConfigurationResolver {
	return &IndexConfigurationResolver{
		configuration: configuration,
	}
}

func (r *IndexConfigurationResolver) Configuration() *string {
	return strPtr(string(r.configuration.Data))
}
