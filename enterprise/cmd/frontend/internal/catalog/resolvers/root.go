package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/catalog"
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
	switch args.Type {
	case "COMPONENT":
		components := dummyComponents(r.db)
		for _, c := range components {
			if c.Name() == args.Name {
				return &gql.CatalogEntityResolver{CatalogEntity: c}, nil
			}
		}

	case "PACKAGE":
		for _, pkg := range catalog.AllPackages() {
			if pkg.Name == args.Name {
				return &gql.CatalogEntityResolver{CatalogEntity: &packageResolver{db: r.db, pkg: pkg}}, nil
			}
		}
	}

	return nil, nil
}

func (r *rootResolver) NodeResolvers() map[string]gql.NodeByIDFunc {
	return map[string]gql.NodeByIDFunc{
		"CatalogComponent": func(ctx context.Context, id graphql.ID) (gql.Node, error) {
			return entityByID(r.db, id), nil
		},
		"Package": func(ctx context.Context, id graphql.ID) (gql.Node, error) {
			return packageByID(r.db, id), nil
		},
		"Group": func(ctx context.Context, id graphql.ID) (gql.Node, error) {
			return groupByID(r.db, id), nil
		},
	}
}
