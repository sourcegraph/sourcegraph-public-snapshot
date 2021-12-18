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

func (r *rootResolver) Component(ctx context.Context, args *gql.ComponentArgs) (gql.ComponentResolver, error) {
	components := dummyComponents(r.db)
	for _, c := range components {
		if c.Name() == args.Name {
			return c, nil
		}
	}
	return nil, nil
}

func (r *rootResolver) Components(ctx context.Context, args *gql.CatalogComponentsArgs) (gql.ComponentConnectionResolver, error) {
	var query string
	if args.Query != nil {
		query = *args.Query
	}
	q := parseQuery(r.db, query)

	var keep []gql.ComponentResolver
	for _, c := range dummyComponents(r.db) {
		if q.matchNode(c) {
			keep = append(keep, c)
		}
	}

	return &componentConnectionResolver{components: keep}, nil
}

func (r *rootResolver) NodeResolvers() map[string]gql.NodeByIDFunc {
	return map[string]gql.NodeByIDFunc{
		"Component": func(ctx context.Context, id graphql.ID) (gql.Node, error) {
			return componentByID(r.db, id), nil
		},
		"Package": func(ctx context.Context, id graphql.ID) (gql.Node, error) {
			return packageByID(r.db, id), nil
		},
		"Group": func(ctx context.Context, id graphql.ID) (gql.Node, error) {
			return groupByID(r.db, id), nil
		},
	}
}
