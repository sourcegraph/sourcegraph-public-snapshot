package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	sgactor "github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *UserResolver) OrganizationMemberships(ctx context.Context) (*organizationMembershipConnectionResolver, error) {
	// 🚨 SECURITY: Only the user and admins are allowed to access the user's
	// organisation memberships.
	if err := auth.CheckSiteAdminOrSameUser(ctx, r.db, r.user.ID); err != nil {
		return nil, err
	}
	memberships, err := r.db.OrgMembers().GetByUserID(ctx, r.user.ID)
	if err != nil {
		return nil, err
	}
	c := organizationMembershipConnectionResolver{nodes: make([]*organizationMembershipResolver, len(memberships))}
	for i, member := range memberships {
		c.nodes[i] = &organizationMembershipResolver{r.db, member}
	}
	return &c, nil
}

func (r *schemaResolver) AutocompleteMembersSearch(ctx context.Context, args *struct {
	Organization graphql.ID
	Query        string
}) ([]*OrgMemberAutocompleteSearchItemResolver, error) {
	actor := sgactor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		return nil, errors.New("no current user")
	}

	orgID, err := UnmarshalOrgID(args.Organization)
	if err != nil {
		return nil, err
	}

	usersMatching, err := r.db.OrgMembers().AutocompleteMembersSearch(ctx, orgID, args.Query)

	if err != nil {
		return nil, err
	}

	var users []*OrgMemberAutocompleteSearchItemResolver
	for _, user := range usersMatching {
		users = append(users, NewOrgMemberAutocompleteSearchItemResolver(r.db, user))
	}

	return users, nil
}

func (r *schemaResolver) OrgMembersSummary(ctx context.Context, args *struct {
	Organization graphql.ID
}) (*OrgMembersSummaryResolver, error) {
	actor := sgactor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		return nil, errors.New("no current user")
	}

	orgID, err := UnmarshalOrgID(args.Organization)
	if err != nil {
		return nil, err
	}

	usersCount, err := r.db.OrgMembers().MemberCount(ctx, orgID)

	if err != nil {
		return nil, err
	}

	pendingInvites, err := r.db.OrgInvitations().Count(ctx, database.OrgInvitationsListOptions{OrgID: orgID})

	if err != nil {
		return nil, err
	}

	var summary = NewOrgMembersSummaryResolver(r.db, orgID, int32(usersCount), int32(pendingInvites))

	return summary, nil
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
	db         database.DB
	membership *types.OrgMembership
}

func (r *organizationMembershipResolver) Organization(ctx context.Context) (*OrgResolver, error) {
	return OrgByIDInt32(ctx, r.db, r.membership.OrgID)
}

func (r *organizationMembershipResolver) User(ctx context.Context) (*UserResolver, error) {
	return UserByIDInt32(ctx, r.db, r.membership.UserID)
}

func (r *organizationMembershipResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.membership.CreatedAt}
}

func (r *organizationMembershipResolver) UpdatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.membership.UpdatedAt}
}

type OrgMemberAutocompleteSearchItemResolver struct {
	db   database.DB
	user *types.OrgMemberAutocompleteSearchItem
}

func (r *OrgMemberAutocompleteSearchItemResolver) ID() graphql.ID {
	return MarshalUserID(r.user.ID)
}

func (r *OrgMemberAutocompleteSearchItemResolver) Username() string {
	return r.user.Username
}

func (r *OrgMemberAutocompleteSearchItemResolver) DisplayName() *string {
	if r.user.DisplayName == "" {
		return nil
	}
	return &r.user.DisplayName
}

func (r *OrgMemberAutocompleteSearchItemResolver) AvatarURL() *string {
	if r.user.AvatarURL == "" {
		return nil
	}
	return &r.user.AvatarURL
}

func (r *OrgMemberAutocompleteSearchItemResolver) InOrg() *bool {
	inOrg := r.user.InOrg > 0
	return &inOrg
}

func NewOrgMemberAutocompleteSearchItemResolver(db database.DB, user *types.OrgMemberAutocompleteSearchItem) *OrgMemberAutocompleteSearchItemResolver {
	return &OrgMemberAutocompleteSearchItemResolver{db: db, user: user}
}

type OrgMembersSummaryResolver struct {
	db           database.DB
	id           int32
	membersCount int32
	invitesCount int32
}

func NewOrgMembersSummaryResolver(db database.DB, orgId int32, membersCount int32, invitesCount int32) *OrgMembersSummaryResolver {
	return &OrgMembersSummaryResolver{
		db:           db,
		id:           orgId,
		membersCount: membersCount,
		invitesCount: invitesCount,
	}
}
func (r *OrgMembersSummaryResolver) ID() graphql.ID {
	return MarshalUserID(r.id)
}

func (r *OrgMembersSummaryResolver) MembersCount() int32 {
	return r.membersCount
}

func (r *OrgMembersSummaryResolver) InvitesCount() int32 {
	return r.invitesCount
}
