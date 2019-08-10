package git

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type FileDiffConnection []graphqlbackend.FileDiff

func (c FileDiffConnection) Nodes(context.Context) ([]graphqlbackend.FileDiff, error) {
	return []graphqlbackend.FileDiff(c), nil
}

func (c FileDiffConnection) TotalCount(context.Context) (int32, error) { return int32(len(c)), nil }

func (c FileDiffConnection) PageInfo(context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}
