package graphqlbackend

import (
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type userTagResolver struct {
	userTag *sourcegraph.UserTag
}

func (u *userTagResolver) ID() int32 {
	return u.userTag.ID
}

func (u *userTagResolver) Name() string {
	return u.userTag.Name
}
