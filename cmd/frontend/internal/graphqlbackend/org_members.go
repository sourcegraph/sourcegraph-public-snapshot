package graphqlbackend

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func (r *userResolver) OrganizationMemberships(ctx context.Context) (*organizationMembershipConnectionResolver, error) {
	memberships, err := db.OrgMembers.GetByUserID(ctx, r.user.ID)
	if err != nil {
		return nil, err
	}
	c := organizationMembershipConnectionResolver{nodes: make([]*organizationMembershipResolver, len(memberships))}
	for i, member := range memberships {
		c.nodes[i] = &organizationMembershipResolver{member}
	}
	return &c, nil
}

type organizationMembershipConnectionResolver struct {
	nodes []*organizationMembershipResolver
}

func (r *organizationMembershipConnectionResolver) Nodes() []*organizationMembershipResolver {
	return r.nodes
}
func (r *organizationMembershipConnectionResolver) TotalCount() int32 { return int32(len(r.nodes)) }
func (r *organizationMembershipConnectionResolver) PageInfo() *pageInfo {
	return &pageInfo{hasNextPage: false}
}

type organizationMembershipResolver struct {
	membership *types.OrgMembership
}

func (r *organizationMembershipResolver) Organization(ctx context.Context) (*orgResolver, error) {
	return orgByIDInt32(ctx, r.membership.OrgID)
}

func (r *organizationMembershipResolver) User(ctx context.Context) (*userResolver, error) {
	return userByIDInt32(ctx, r.membership.UserID)
}

func (r *organizationMembershipResolver) CreatedAt() string {
	return r.membership.CreatedAt.Format(time.RFC3339)
}

func (r *organizationMembershipResolver) UpdatedAt() string {
	return r.membership.UpdatedAt.Format(time.RFC3339)
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
		email, _, err := db.UserEmails.GetPrimaryEmail(ctx, m.UserID)
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
