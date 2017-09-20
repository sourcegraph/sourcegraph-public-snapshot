package graphqlbackend

import (
	"context"
	"time"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

type orgMemberResolver struct {
	org    *sourcegraph.Org
	member *sourcegraph.OrgMember
}

func (m *orgMemberResolver) ID() int32 {
	return m.member.ID
}

func (m *orgMemberResolver) Org(ctx context.Context) (*orgResolver, error) {
	if m.org == nil {
		var err error
		m.org, err = localstore.Orgs.GetByID(ctx, m.member.OrgID)
		if err != nil {
			return nil, err
		}
	}
	return &orgResolver{m.org}, nil
}

func (m *orgMemberResolver) UserID() string {
	return m.member.UserID
}

func (m *orgMemberResolver) Email() string {
	return m.member.Email
}

func (m *orgMemberResolver) Username() string {
	return m.member.Username
}

func (m *orgMemberResolver) DisplayName() string {
	return m.member.DisplayName
}

func (m *orgMemberResolver) AvatarURL() *string {
	return m.member.AvatarURL
}

func (m *orgMemberResolver) CreatedAt() string {
	return m.member.CreatedAt.Format(time.RFC3339) // ISO
}

func (m *orgMemberResolver) UpdatedAt() string {
	return m.member.UpdatedAt.Format(time.RFC3339) // ISO
}
