package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

var _ graphqlbackend.InsightDirtyQueryResolver = &insightDirtyQueryResolver{}

type insightDirtyQueryResolver struct {
	data *types.DirtyQueryAggregate
}

func (i *insightDirtyQueryResolver) Reason(ctx context.Context) string {
	return i.data.Reason
}

func (i *insightDirtyQueryResolver) Time(ctx context.Context) gqlutil.DateTime {
	return gqlutil.DateTime{Time: i.data.ForTime}
}

func (i *insightDirtyQueryResolver) Count(ctx context.Context) int32 {
	return int32(i.data.Count)
}
