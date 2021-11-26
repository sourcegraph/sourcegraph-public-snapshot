package resolvers

import (
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/catalog"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

// TODO(sqs): dummy data
func dummyComponents(db database.DB) []*componentResolver {
	components := catalog.Components()
	return componentResolvers(db, components)
}

func componentResolvers(db database.DB, components []catalog.Component) []*componentResolver {
	resolvers := make([]*componentResolver, len(components))
	for i, c := range components {
		resolvers[i] = &componentResolver{
			component: c,
			db:        db,
		}
	}
	return resolvers
}

func componentResolversGQLIface(db database.DB, components []catalog.Component) []gql.ComponentResolver {
	resolvers := make([]gql.ComponentResolver, len(components))
	for i, c := range components {
		resolvers[i] = &componentResolver{
			component: c,
			db:        db,
		}
	}
	return resolvers
}
