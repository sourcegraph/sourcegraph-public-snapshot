package resolvers

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/catalog"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type catalogResolver struct {
	db database.DB
}

func (r *catalogResolver) Entities(ctx context.Context, args *gql.CatalogEntitiesArgs) (gql.CatalogEntityConnectionResolver, error) {
	var query string
	if args.Query != nil {
		query = *args.Query
	}
	q := parseQuery(r.db, query)

	var keep []gql.CatalogEntity

	for _, c := range dummyComponents(r.db) {
		if q.matchNode(c) {
			keep = append(keep, c)
		}
	}

	for _, p := range catalog.AllPackages() {
		if q.matchPackage(p) {
			keep = append(keep, &packageResolver{db: r.db, pkg: p})
		}
	}

	return &catalogEntityConnectionResolver{entities: wrapInCatalogEntityInterfaceType(keep)}, nil
}
