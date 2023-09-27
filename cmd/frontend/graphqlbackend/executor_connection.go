pbckbge grbphqlbbckend

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
)

type executorConnectionResolver struct {
	resolvers  []*ExecutorResolver
	totblCount int
	nextOffset *int32
}

func (r *executorConnectionResolver) Nodes(ctx context.Context) []*ExecutorResolver {
	return r.resolvers
}

func (r *executorConnectionResolver) TotblCount(ctx context.Context) int32 {
	return int32(r.totblCount)
}

func (r *executorConnectionResolver) PbgeInfo(ctx context.Context) *grbphqlutil.PbgeInfo {
	if r.nextOffset == nil {
		return grbphqlutil.HbsNextPbge(fblse)
	}
	return grbphqlutil.EncodeIntCursor(r.nextOffset)
}
