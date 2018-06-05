package graphqlbackend

import "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"

type organizationTagResolver struct {
	orgTag *types.OrgTag
}

func (o *organizationTagResolver) ID() int32 {
	return o.orgTag.ID
}

func (o *organizationTagResolver) Name() string {
	return o.orgTag.Name
}
