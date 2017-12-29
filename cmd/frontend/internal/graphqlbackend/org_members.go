package graphqlbackend

import (
	"context"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
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

func (m *orgMemberResolver) User(ctx context.Context) (*userResolver, error) {
	if m.user == nil {
		var err error
		m.user, err = localstore.Users.GetByAuthID(ctx, m.member.UserID)
		if err != nil {
			return nil, err
		}
	}
	return &userResolver{m.user}, nil
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

var mockAllEmailsForOrg func(ctx context.Context, orgID int32, excludeByUserID []string) ([]string, error)

func allEmailsForOrg(ctx context.Context, orgID int32, excludeByUserID []string) ([]string, error) {
	if mockAllEmailsForOrg != nil {
		return mockAllEmailsForOrg(ctx, orgID, excludeByUserID)
	}

	members, err := store.OrgMembers.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	exclude := make(map[string]interface{})
	for _, id := range excludeByUserID {
		exclude[id] = struct{}{}
	}
	emails := []string{}
	for _, m := range members {
		if _, ok := exclude[m.UserID]; ok {
			continue
		}
		user, err := store.Users.GetByAuthID(ctx, m.UserID)
		if err != nil {
			// This shouldn't happen, but we don't want to prevent the notification,
			// so swallow the error.
			log15.Error("get user", "uid", m.UserID, "error", err)
			continue
		}
		emails = append(emails, user.Email)
	}
	return emails, nil
}
