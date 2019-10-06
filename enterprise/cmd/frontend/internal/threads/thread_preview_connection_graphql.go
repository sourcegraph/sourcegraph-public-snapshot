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

func (c ConstThreadOrThreadPreviewConnection) Filters(context.Context) (graphqlbackend.ThreadConnectionFilters, error) {
	return newThreadConnectionFiltersFromConst(c), nil
}

func ToThreadOrThreadPreviews(threads []graphqlbackend.Thread, threadPreviews []graphqlbackend.ThreadPreview) []graphqlbackend.ToThreadOrThreadPreview {
	v := make([]graphqlbackend.ToThreadOrThreadPreview, len(threads)+len(threadPreviews))
	for i, t := range threads {
		v[i] = graphqlbackend.ToThreadOrThreadPreview{Thread: t}
	}
	for i, t := range threadPreviews {
		v[len(threads)+i] = graphqlbackend.ToThreadOrThreadPreview{ThreadPreview: t}
	}
	return v
}
