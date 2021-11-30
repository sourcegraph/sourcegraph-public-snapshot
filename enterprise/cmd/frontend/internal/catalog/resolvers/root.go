package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func NewRootResolver(db database.DB) gql.CatalogRootResolver {
	return &rootResolver{db: db}
}

type rootResolver struct {
	db database.DB
}

func (r *rootResolver) Catalog(context.Context) (gql.CatalogResolver, error) {
	return &catalogResolver{db: r.db}, nil
}

func (r *rootResolver) CatalogEntity(ctx context.Context, args *gql.CatalogEntityArgs) (*gql.CatalogEntityResolver, error) {
	components := dummyData(r.db)
	for _, c := range components {
		if c.Name() == args.Name {
			return &gql.CatalogEntityResolver{c}, nil
		}
	}
	return nil, nil
}

func (r *rootResolver) NodeResolvers() map[string]gql.NodeByIDFunc {
	return map[string]gql.NodeByIDFunc{
		"CatalogComponent": func(ctx context.Context, id graphql.ID) (gql.Node, error) {
			components := dummyData(r.db)
			for _, c := range components {
				if c.ID() == id {
					return c, nil
				}
			}
			return nil, nil
		},
	}
}
