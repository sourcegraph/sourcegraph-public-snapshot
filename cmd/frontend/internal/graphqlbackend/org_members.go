package graphqlbackend

import (
	"time"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type orgMemberResolver struct {
	member *sourcegraph.OrgMember
}

func (m *orgMemberResolver) ID() int32 {
	return m.member.ID
}

func (m *orgMemberResolver) OrgID() int32 {
	return m.member.OrgID
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

func (m *orgMemberResolver) CreatedAt() string {
	return m.member.CreatedAt.Format(time.RFC3339) // ISO
}

func (m *orgMemberResolver) UpdatedAt() string {
	return m.member.UpdatedAt.Format(time.RFC3339) // ISO
}
