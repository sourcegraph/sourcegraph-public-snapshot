package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type executorConnectionResolver struct {
	resolvers  []*ExecutorResolver
	totalCount int
	nextOffset *int
}

func (r *executorConnectionResolver) Nodes(ctx context.Context) []*ExecutorResolver {
	return r.resolvers
}

func (r *executorConnectionResolver) TotalCount(ctx context.Context) int32 {
	return int32(r.totalCount)
}

func (r *executorConnectionResolver) PageInfo(ctx context.Context) *graphqlutil.PageInfo {
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
