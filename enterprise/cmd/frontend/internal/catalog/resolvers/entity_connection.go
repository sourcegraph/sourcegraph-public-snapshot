package resolvers

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type catalogEntityConnectionResolver struct {
	entities []*gql.CatalogEntityResolver
}

func (r *catalogEntityConnectionResolver) Nodes(ctx context.Context) ([]*gql.CatalogEntityResolver, error) {
	return r.entities, nil
}

func (r *catalogEntityConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return int32(len(r.entities)), nil
}

func (r *catalogEntityConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil // TODO(sqs)
}
