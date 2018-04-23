package graphqlbackend

import "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"

type userTagResolver struct {
	userTag *types.UserTag
}

func (u *userTagResolver) ID() int32 {
	return u.userTag.ID
}

func (u *userTagResolver) Name() string {
	return u.userTag.Name
}
