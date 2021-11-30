package resolvers

import (
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func wrapInCatalogEntityInterfaceType(entities []gql.CatalogEntity) []*gql.CatalogEntityResolver {
	resolvers := make([]*gql.CatalogEntityResolver, len(entities))
	for i, e := range entities {
		resolvers[i] = &gql.CatalogEntityResolver{e}
	}
	return resolvers
}
