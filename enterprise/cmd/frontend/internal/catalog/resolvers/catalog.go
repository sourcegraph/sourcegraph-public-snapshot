package resolvers

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type catalogResolver struct {
	db database.DB
}

func (r *catalogResolver) Entities(ctx context.Context, args *gql.CatalogEntitiesArgs) (gql.CatalogEntityConnectionResolver, error) {
	components := dummyData(r.db)

	var query string
	if args.Query != nil {
		query = *args.Query
	}
	match := getQueryMatcher(query)

	var keep []gql.CatalogEntity
	for _, c := range components {
		if match(c) {
			keep = append(keep, c)
		}
	}

	return &catalogEntityConnectionResolver{entities: wrapInCatalogEntityInterfaceType(keep)}, nil
}
