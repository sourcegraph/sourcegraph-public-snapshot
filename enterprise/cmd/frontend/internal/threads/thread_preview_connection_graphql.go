package threads

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type ConstThreadOrThreadPreviewConnection []graphqlbackend.ToThreadOrThreadPreview

func (c ConstThreadOrThreadPreviewConnection) Nodes(context.Context) ([]graphqlbackend.ToThreadOrThreadPreview, error) {
	return []graphqlbackend.ToThreadOrThreadPreview(c), nil
}

func (c ConstThreadOrThreadPreviewConnection) TotalCount(context.Context) (int32, error) {
	return int32(len(c)), nil
}

func (c ConstThreadOrThreadPreviewConnection) PageInfo(context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}
