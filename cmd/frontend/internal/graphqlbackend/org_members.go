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
	user   *sourcegraph.User
}

type orgInviteResolver struct {
	emailVerified bool
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

// DEPRECATED (use embedded User instead).
func (m *orgMemberResolver) Username() string {
	user, err := localstore.Users.GetByAuth0ID(m.UserID())
	if err != nil {
		return ""
	}
	return user.Username
}

// DEPRECATED (use embedded User instead).
func (m *orgMemberResolver) Email() string {
	user, err := localstore.Users.GetByAuth0ID(m.UserID())
	if err != nil {
		return ""
	}
	return user.Email
}

// DEPRECATED (use embedded User instead).
func (m *orgMemberResolver) DisplayName() string {
	user, err := localstore.Users.GetByAuth0ID(m.UserID())
	if err != nil {
		return ""
	}
	return user.DisplayName
}

// DEPRECATED (use embedded User instead).
func (m *orgMemberResolver) AvatarURL() *string {
	user, err := localstore.Users.GetByAuth0ID(m.UserID())
	if err != nil {
		return nil
	}
	return user.AvatarURL
}

func (m *orgMemberResolver) User(ctx context.Context) (*userResolver, error) {
	if m.user == nil {
		var err error
		m.user, err = localstore.Users.GetByAuth0ID(m.UserID())
		if err != nil {
			return nil, err
		}
	}
	return &userResolver{m.user, nil}, nil
}

func (m *orgMemberResolver) CreatedAt() string {
	return m.member.CreatedAt.Format(time.RFC3339) // ISO
}

func (m *orgMemberResolver) UpdatedAt() string {
	return m.member.UpdatedAt.Format(time.RFC3339) // ISO
}

func (i *orgInviteResolver) EmailVerified() bool {
	return i.emailVerified
}
