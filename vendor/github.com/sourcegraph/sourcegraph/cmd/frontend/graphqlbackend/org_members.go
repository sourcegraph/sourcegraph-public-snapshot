package graphqlbackend

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
)

func (r *UserResolver) OrganizationMemberships(ctx context.Context) (*organizationMembershipConnectionResolver, error) {
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
func (r *organizationMembershipConnectionResolver) PageInfo() *PageInfo {
	return &PageInfo{hasNextPage: false}
}

type organizationMembershipResolver struct {
	membership *types.OrgMembership
}

func (r *organizationMembershipResolver) Organization(ctx context.Context) (*orgResolver, error) {
	return orgByIDInt32(ctx, r.membership.OrgID)
}

func (r *organizationMembershipResolver) User(ctx context.Context) (*UserResolver, error) {
	return UserByIDInt32(ctx, r.membership.UserID)
}

func (r *organizationMembershipResolver) CreatedAt() string {
	return r.membership.CreatedAt.Format(time.RFC3339)
}

func (r *organizationMembershipResolver) UpdatedAt() string {
	return r.membership.UpdatedAt.Format(time.RFC3339)
}
