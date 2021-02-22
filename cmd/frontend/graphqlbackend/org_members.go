package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func (r *UserResolver) OrganizationMemberships(ctx context.Context) (*organizationMembershipConnectionResolver, error) {
	memberships, err := database.OrgMembers(r.db).GetByUserID(ctx, r.user.ID)
	if err != nil {
		return nil, err
	}
	c := organizationMembershipConnectionResolver{nodes: make([]*organizationMembershipResolver, len(memberships))}
	for i, member := range memberships {
		c.nodes[i] = &organizationMembershipResolver{r.db, member}
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
func (r *organizationMembershipConnectionResolver) PageInfo() *graphqlutil.PageInfo {
	return graphqlutil.HasNextPage(false)
}

type organizationMembershipResolver struct {
	db         dbutil.DB
	membership *types.OrgMembership
}

func (r *organizationMembershipResolver) Organization(ctx context.Context) (*OrgResolver, error) {
	return OrgByIDInt32(ctx, r.db, r.membership.OrgID)
}

func (r *organizationMembershipResolver) User(ctx context.Context) (*UserResolver, error) {
	return UserByIDInt32(ctx, r.db, r.membership.UserID)
}

func (r *organizationMembershipResolver) CreatedAt() DateTime {
	return DateTime{Time: r.membership.CreatedAt}
}

func (r *organizationMembershipResolver) UpdatedAt() DateTime {
	return DateTime{Time: r.membership.UpdatedAt}
}
