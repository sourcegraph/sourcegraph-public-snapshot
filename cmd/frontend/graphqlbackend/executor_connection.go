package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type executorConnectionResolver struct {
	resolvers  []*ExecutorResolver
	totalCount int
	nextOffset *int32
}

func (r *executorConnectionResolver) Nodes(ctx context.Context) []*ExecutorResolver {
	return r.resolvers
}

func (r *executorConnectionResolver) TotalCount(ctx context.Context) int32 {
	return int32(r.totalCount)
}

func (r *executorConnectionResolver) PageInfo(ctx context.Context) *graphqlutil.PageInfo {
	if r.nextOffset == nil {
		return graphqlutil.HasNextPage(false)
	}
	return graphqlutil.EncodeIntCursor(r.nextOffset)
}
