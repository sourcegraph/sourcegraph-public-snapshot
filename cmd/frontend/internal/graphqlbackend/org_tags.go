package graphqlbackend

import "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"

type orgTagResolver struct {
	orgTag *types.OrgTag
}

func (o *orgTagResolver) ID() int32 {
	return o.orgTag.ID
}

func (o *orgTagResolver) Name() string {
	return o.orgTag.Name
}
