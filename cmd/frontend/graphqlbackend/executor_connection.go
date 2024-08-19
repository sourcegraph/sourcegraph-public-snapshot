package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
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

func (r *executorConnectionResolver) PageInfo(ctx context.Context) *gqlutil.PageInfo {
	if r.nextOffset == nil {
		return gqlutil.HasNextPage(false)
	}
	return gqlutil.EncodeIntCursor(r.nextOffset)
}
