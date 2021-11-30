package resolvers

import (
	"context"
	"strings"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type catalogResolver struct {
	db database.DB
}

func (r *catalogResolver) Entities(ctx context.Context, args *gql.CatalogEntitiesArgs) (gql.CatalogEntityConnectionResolver, error) {
	components := dummyData(r.db)

	var keep []gql.CatalogEntity
	for _, c := range components {
		if args.Query == nil || strings.Contains(c.component.Name, *args.Query) {
			keep = append(keep, c)
		}
	}

	return &catalogEntityConnectionResolver{entities: wrapInCatalogEntityInterfaceType(keep)}, nil
}
