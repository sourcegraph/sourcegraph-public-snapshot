package resolvers

import (
	"github.com/graph-gophers/graphql-go"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func entityByID(db database.DB, id graphql.ID) *catalogComponentResolver {
	components := dummyComponents(db)
	for _, c := range components {
		if c.ID() == id {
			return c
		}
	}
	return nil
}

func wrapInCatalogEntityInterfaceType(entities []gql.CatalogEntity) []*gql.CatalogEntityResolver {
	resolvers := make([]*gql.CatalogEntityResolver, len(entities))
	for i, e := range entities {
		resolvers[i] = &gql.CatalogEntityResolver{e}
	}
	return resolvers
}
