package graphqlbackend

import (
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type orgTagResolver struct {
	orgTag *sourcegraph.OrgTag
}

func (o *orgTagResolver) ID() int32 {
	return o.orgTag.ID
}

func (o *orgTagResolver) Name() string {
	return o.orgTag.Name
}
