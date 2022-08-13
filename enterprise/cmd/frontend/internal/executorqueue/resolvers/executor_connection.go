package resolvers

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type executorConnectionResolver struct {
	resolvers  []graphqlbackend.ExecutorResolver
	totalCount int
	nextOffset *int
}

var _ graphqlbackend.ExecutorConnectionResolver = &executorConnectionResolver{}

func (r *executorConnectionResolver) Nodes() []graphqlbackend.ExecutorResolver {
	return r.resolvers
}

func (r *executorConnectionResolver) TotalCount() int32 {
	return int32(r.totalCount)
}

func (r *executorConnectionResolver) PageInfo() *graphqlutil.PageInfo {
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
