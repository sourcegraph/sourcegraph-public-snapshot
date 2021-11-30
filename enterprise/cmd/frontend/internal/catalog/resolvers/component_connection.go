package resolvers

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type catalogComponentConnectionResolver struct {
	components []gql.CatalogComponentResolver
}

func (r *catalogComponentConnectionResolver) Nodes(ctx context.Context) ([]gql.CatalogComponentResolver, error) {
	return r.components, nil
}

func (r *catalogComponentConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return int32(len(r.components)), nil
}

func (r *catalogComponentConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil // TODO(sqs)
}
