package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type ExecutorConnection struct {
	resolvers  []*ExecutorResolver
	totalCount int
	nextOffset *int
}

func NewExecutorConnection(resolvers []*ExecutorResolver, totalCount int, nextOffset *int) *ExecutorConnection {
	return &ExecutorConnection{
		resolvers:  resolvers,
		totalCount: totalCount,
		nextOffset: nextOffset,
	}
}

func (r *ExecutorConnection) Nodes(ctx context.Context) []*ExecutorResolver {
	return r.resolvers
}

func (r *ExecutorConnection) TotalCount(ctx context.Context) int32 {
	return int32(r.totalCount)
}

func (r *ExecutorConnection) PageInfo(ctx context.Context) *graphqlutil.PageInfo {
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
