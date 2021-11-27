package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func NewRootResolver(db database.DB) graphqlbackend.CatalogRootResolver {
	return &rootResolver{db: db}
}

type rootResolver struct {
	db database.DB
}

func (r *rootResolver) Catalog(context.Context) (graphqlbackend.CatalogResolver, error) {
	return &catalogResolver{}, nil
}

type catalogResolver struct{}

func (r *catalogResolver) Foo() ([]string, error) {
	return []string{"alice", "bob", "carol"}, nil
}
