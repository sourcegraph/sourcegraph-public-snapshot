package resolvers

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type componentConnectionResolver struct {
	components []gql.ComponentResolver
}

func (r *componentConnectionResolver) Nodes(ctx context.Context) ([]gql.ComponentResolver, error) {
	return r.components, nil
}

func (r *componentConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return int32(len(r.components)), nil
}

func (r *componentConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil // TODO(sqs)
}
