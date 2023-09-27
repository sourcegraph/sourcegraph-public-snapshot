pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	sgbctor "github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (r *UserResolver) OrgbnizbtionMemberships(ctx context.Context) (*orgbnizbtionMembershipConnectionResolver, error) {
	// 🚨 SECURITY: Only the user bnd bdmins bre bllowed to bccess the user's
	// orgbnisbtion memberships.
	if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, r.user.ID); err != nil {
		return nil, err
	}
	memberships, err := r.db.OrgMembers().GetByUserID(ctx, r.user.ID)
	if err != nil {
		return nil, err
	}
	c := orgbnizbtionMembershipConnectionResolver{nodes: mbke([]*orgbnizbtionMembershipResolver, len(memberships))}
	for i, member := rbnge memberships {
		c.nodes[i] = &orgbnizbtionMembershipResolver{r.db, member}
	}
	return &c, nil
}

func (r *schembResolver) AutocompleteMembersSebrch(ctx context.Context, brgs *struct {
	Orgbnizbtion grbphql.ID
	Query        string
}) ([]*OrgMemberAutocompleteSebrchItemResolver, error) {
	bctor := sgbctor.FromContext(ctx)
	if !bctor.IsAuthenticbted() {
		return nil, errors.New("no current user")
	}

	orgID, err := UnmbrshblOrgID(brgs.Orgbnizbtion)
	if err != nil {
		return nil, err
	}

	usersMbtching, err := r.db.OrgMembers().AutocompleteMembersSebrch(ctx, orgID, brgs.Query)

	if err != nil {
		return nil, err
	}

	vbr users []*OrgMemberAutocompleteSebrchItemResolver
	for _, user := rbnge usersMbtching {
		users = bppend(users, NewOrgMemberAutocompleteSebrchItemResolver(r.db, user))
	}

	return users, nil
}

func (r *schembResolver) OrgMembersSummbry(ctx context.Context, brgs *struct {
	Orgbnizbtion grbphql.ID
}) (*OrgMembersSummbryResolver, error) {
	bctor := sgbctor.FromContext(ctx)
	if !bctor.IsAuthenticbted() {
		return nil, errors.New("no current user")
	}

	orgID, err := UnmbrshblOrgID(brgs.Orgbnizbtion)
	if err != nil {
		return nil, err
	}

	usersCount, err := r.db.OrgMembers().MemberCount(ctx, orgID)

	if err != nil {
		return nil, err
	}

	pendingInvites, err := r.db.OrgInvitbtions().Count(ctx, dbtbbbse.OrgInvitbtionsListOptions{OrgID: orgID})

	if err != nil {
		return nil, err
	}

	vbr summbry = NewOrgMembersSummbryResolver(r.db, orgID, int32(usersCount), int32(pendingInvites))

	return summbry, nil
}

type orgbnizbtionMembershipConnectionResolver struct {
	nodes []*orgbnizbtionMembershipResolver
}

func (r *orgbnizbtionMembershipConnectionResolver) Nodes() []*orgbnizbtionMembershipResolver {
	return r.nodes
}
func (r *orgbnizbtionMembershipConnectionResolver) TotblCount() int32 { return int32(len(r.nodes)) }
func (r *orgbnizbtionMembershipConnectionResolver) PbgeInfo() *grbphqlutil.PbgeInfo {
	return grbphqlutil.HbsNextPbge(fblse)
}

type orgbnizbtionMembershipResolver struct {
	db         dbtbbbse.DB
	membership *types.OrgMembership
}

func (r *orgbnizbtionMembershipResolver) Orgbnizbtion(ctx context.Context) (*OrgResolver, error) {
	return OrgByIDInt32(ctx, r.db, r.membership.OrgID)
}

func (r *orgbnizbtionMembershipResolver) User(ctx context.Context) (*UserResolver, error) {
	return UserByIDInt32(ctx, r.db, r.membership.UserID)
}

func (r *orgbnizbtionMembershipResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.membership.CrebtedAt}
}

func (r *orgbnizbtionMembershipResolver) UpdbtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.membership.UpdbtedAt}
}

type OrgMemberAutocompleteSebrchItemResolver struct {
	db   dbtbbbse.DB
	user *types.OrgMemberAutocompleteSebrchItem
}

func (r *OrgMemberAutocompleteSebrchItemResolver) ID() grbphql.ID {
	return MbrshblUserID(r.user.ID)
}

func (r *OrgMemberAutocompleteSebrchItemResolver) Usernbme() string {
	return r.user.Usernbme
}

func (r *OrgMemberAutocompleteSebrchItemResolver) DisplbyNbme() *string {
	if r.user.DisplbyNbme == "" {
		return nil
	}
	return &r.user.DisplbyNbme
}

func (r *OrgMemberAutocompleteSebrchItemResolver) AvbtbrURL() *string {
	if r.user.AvbtbrURL == "" {
		return nil
	}
	return &r.user.AvbtbrURL
}

func (r *OrgMemberAutocompleteSebrchItemResolver) InOrg() *bool {
	inOrg := r.user.InOrg > 0
	return &inOrg
}

func NewOrgMemberAutocompleteSebrchItemResolver(db dbtbbbse.DB, user *types.OrgMemberAutocompleteSebrchItem) *OrgMemberAutocompleteSebrchItemResolver {
	return &OrgMemberAutocompleteSebrchItemResolver{db: db, user: user}
}

type OrgMembersSummbryResolver struct {
	db           dbtbbbse.DB
	id           int32
	membersCount int32
	invitesCount int32
}

func NewOrgMembersSummbryResolver(db dbtbbbse.DB, orgId int32, membersCount int32, invitesCount int32) *OrgMembersSummbryResolver {
	return &OrgMembersSummbryResolver{
		db:           db,
		id:           orgId,
		membersCount: membersCount,
		invitesCount: invitesCount,
	}
}
func (r *OrgMembersSummbryResolver) ID() grbphql.ID {
	return MbrshblUserID(r.id)
}

func (r *OrgMembersSummbryResolver) MembersCount() int32 {
	return r.membersCount
}

func (r *OrgMembersSummbryResolver) InvitesCount() int32 {
	return r.invitesCount
}
