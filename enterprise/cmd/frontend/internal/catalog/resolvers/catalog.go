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

func (r *catalogResolver) Components(ctx context.Context, args *gql.CatalogComponentsArgs) (gql.CatalogComponentConnectionResolver, error) {
	components := dummyData(r.db)

	var keep []gql.CatalogComponentResolver
	for _, c := range components {
		if args.Query == nil || strings.Contains(c.component.Name, *args.Query) {
			keep = append(keep, c)
		}
	}

	return &catalogComponentConnectionResolver{
		components: keep,
	}, nil
}
