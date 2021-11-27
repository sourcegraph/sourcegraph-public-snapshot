package resolvers

import (
	"context"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
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

func (r *rootResolver) NodeResolvers() map[string]graphqlbackend.NodeByIDFunc {
	return map[string]graphqlbackend.NodeByIDFunc{
		"CatalogComponent": func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
			panic("TODO(sqs)")
		},
	}
}

type catalogResolver struct{}

func (r *catalogResolver) Components(ctx context.Context, args *graphqlbackend.CatalogComponentsArgs) (graphqlbackend.CatalogComponentConnectionResolver, error) {
	components := []*catalogComponentResolver{
		{name: "aaa"},
		{name: "bbb"},
		{name: "ccc"},
	}

	var keep []graphqlbackend.CatalogComponentResolver
	for _, c := range components {
		if args.Query == nil || strings.Contains(c.name, *args.Query) {
			keep = append(keep, c)
		}
	}

	return &catalogComponentConnectionResolver{
		components: keep,
	}, nil
}

type catalogComponentConnectionResolver struct {
	components []graphqlbackend.CatalogComponentResolver
}

func (r *catalogComponentConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.CatalogComponentResolver, error) {
	return r.components, nil
}

func (r *catalogComponentConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return int32(len(r.components)), nil
}

func (r *catalogComponentConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil // TODO(sqs)
}

type catalogComponentResolver struct {
	name string
}

func (r *catalogComponentResolver) ID() graphql.ID {
	return graphql.ID(r.name) // TODO(sqs)
}

func (r *catalogComponentResolver) Name() string {
	return r.name
}

func (r *catalogComponentResolver) Xyz123() string {
	return r.name
}
