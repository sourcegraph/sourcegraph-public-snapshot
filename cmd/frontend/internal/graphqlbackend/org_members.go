package graphqlbackend

import (
	"context"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
)

func (r *userResolver) OrgMemberships(ctx context.Context) ([]*orgMemberResolver, error) {
	members, err := db.OrgMembers.GetByUserID(ctx, r.user.ID)
	if err != nil {
		return nil, err
	}
	orgMemberResolvers := []*orgMemberResolver{}
	for _, member := range members {
		orgMemberResolvers = append(orgMemberResolvers, &orgMemberResolver{nil, member, nil})
	}
	return orgMemberResolvers, nil
}

type orgMemberResolver struct {
	org    *types.Org
	member *types.OrgMembership
	user   *types.User
}

func (m *orgMemberResolver) ID() int32 {
	return m.member.ID
}

func (m *orgMemberResolver) Org(ctx context.Context) (*orgResolver, error) {
	if m.org == nil {
		var err error
		m.org, err = db.Orgs.GetByID(ctx, m.member.OrgID)
		if err != nil {
			return nil, err
		}
	}
	return &orgResolver{m.org}, nil
}

func (m *orgMemberResolver) User(ctx context.Context) (*userResolver, error) {
	if m.user == nil {
		var err error
		m.user, err = db.Users.GetByID(ctx, m.member.UserID)
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

var mockAllEmailsForOrg func(ctx context.Context, orgID int32, excludeByUserID []int32) ([]string, error)

func allEmailsForOrg(ctx context.Context, orgID int32, excludeByUserID []int32) ([]string, error) {
	if mockAllEmailsForOrg != nil {
		return mockAllEmailsForOrg(ctx, orgID, excludeByUserID)
	}

	members, err := db.OrgMembers.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	exclude := make(map[int32]interface{})
	for _, id := range excludeByUserID {
		exclude[id] = struct{}{}
	}
	emails := []string{}
	for _, m := range members {
		if _, ok := exclude[m.UserID]; ok {
			continue
		}
		email, _, err := db.UserEmails.GetEmail(ctx, m.UserID)
		if err != nil {
			// This shouldn't happen, but we don't want to prevent the notification,
			// so swallow the error.
			log15.Error("get user", "uid", m.UserID, "error", err)
			continue
		}
		emails = append(emails, email)
	}
	return emails, nil
}
