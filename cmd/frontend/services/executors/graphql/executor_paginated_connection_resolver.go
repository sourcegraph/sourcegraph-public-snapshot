package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type ExecutorPaginatedResolver struct {
	resolvers  []*ExecutorResolver
	totalCount int
	nextOffset *int
}

func NewExecutorPaginatedConnection(resolvers []*ExecutorResolver, totalCount int, nextOffset *int) *ExecutorPaginatedResolver {
	return &ExecutorPaginatedResolver{
		resolvers:  resolvers,
		totalCount: totalCount,
		nextOffset: nextOffset,
	}
}

func (r *ExecutorPaginatedResolver) Nodes(ctx context.Context) []*ExecutorResolver {
	return r.resolvers
}

func (r *ExecutorPaginatedResolver) TotalCount(ctx context.Context) int32 {
	return int32(r.totalCount)
}

func (r *ExecutorPaginatedResolver) PageInfo(ctx context.Context) *graphqlutil.PageInfo {
	return graphqlutil.EncodeIntCursor(toInt32(r.nextOffset))
}

// toInt32 translates the given int pointer into an int32 pointer.
func toInt32(val *int) *int32 {
	if val == nil {
		return nil
	}

	v := int32(*val)
	return &v
}
