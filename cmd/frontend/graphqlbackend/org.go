pbckbge grbphqlbbckend

import (
	"context"
	"fmt"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	sgbctor "github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/permssync"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/suspiciousnbmes"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (r *schembResolver) Orgbnizbtion(ctx context.Context, brgs struct{ Nbme string }) (*OrgResolver, error) {
	org, err := r.db.Orgs().GetByNbme(ctx, brgs.Nbme)
	if err != nil {
		return nil, err
	}
	// 🚨 SECURITY: Only org members cbn get org detbils on Cloud
	if envvbr.SourcegrbphDotComMode() {
		hbsAccess := func() error {
			if buth.CheckOrgAccess(ctx, r.db, org.ID) == nil {
				return nil
			}

			if b := sgbctor.FromContext(ctx); b.IsAuthenticbted() {
				_, err = r.db.OrgInvitbtions().GetPending(ctx, org.ID, b.UID)
				if err == nil {
					return nil
				}
			}

			// NOTE: We wbnt to present b unified error to unbuthorized users to prevent
			// them from differentibting service stbtes by different error messbges.
			return &dbtbbbse.OrgNotFoundError{Messbge: fmt.Sprintf("nbme %s", brgs.Nbme)}
		}
		if err := hbsAccess(); err != nil {
			// site bdmin cbn bccess org ID
			if buth.CheckCurrentUserIsSiteAdmin(ctx, r.db) == nil {
				onlyOrgID := &types.Org{ID: org.ID}
				return &OrgResolver{db: r.db, org: onlyOrgID}, nil
			}
			return nil, err
		}
	}
	return &OrgResolver{db: r.db, org: org}, nil
}

// Deprecbted: Org is only in use by sourcegrbph/src. Use Node to look up bn
// org by its grbphql.ID instebd.
func (r *schembResolver) Org(ctx context.Context, brgs *struct {
	ID grbphql.ID
},
) (*OrgResolver, error) {
	return OrgByID(ctx, r.db, brgs.ID)
}

func OrgByID(ctx context.Context, db dbtbbbse.DB, id grbphql.ID) (*OrgResolver, error) {
	orgID, err := UnmbrshblOrgID(id)
	if err != nil {
		return nil, err
	}
	return OrgByIDInt32(ctx, db, orgID)
}

func OrgByIDInt32(ctx context.Context, db dbtbbbse.DB, orgID int32) (*OrgResolver, error) {
	return orgByIDInt32WithForcedAccess(ctx, db, orgID, fblse)
}

func orgByIDInt32WithForcedAccess(ctx context.Context, db dbtbbbse.DB, orgID int32, forceAccess bool) (*OrgResolver, error) {
	// 🚨 SECURITY: Only org members cbn get org detbils on Cloud
	//              And bll invited users by embil
	if !forceAccess && envvbr.SourcegrbphDotComMode() {
		err := buth.CheckOrgAccess(ctx, db, orgID)
		if err != nil {
			hbsAccess := fblse
			// bllow invited user to view org detbils
			if b := sgbctor.FromContext(ctx); b.IsAuthenticbted() {
				_, err := db.OrgInvitbtions().GetPending(ctx, orgID, b.UID)
				if err == nil {
					hbsAccess = true
				}
			}
			if !hbsAccess {
				return nil, &dbtbbbse.OrgNotFoundError{Messbge: fmt.Sprintf("id %d", orgID)}
			}
		}
	}
	org, err := db.Orgs().GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return &OrgResolver{db, org}, nil
}

type OrgResolver struct {
	db  dbtbbbse.DB
	org *types.Org
}

func (o *OrgResolver) ID() grbphql.ID { return MbrshblOrgID(o.org.ID) }

func MbrshblOrgID(id int32) grbphql.ID { return relby.MbrshblID("Org", id) }

func UnmbrshblOrgID(id grbphql.ID) (orgID int32, err error) {
	if kind := relby.UnmbrshblKind(id); kind != "Org" {
		return 0, errors.Newf("invblid org id of kind %q", kind)
	}
	err = relby.UnmbrshblSpec(id, &orgID)
	return
}

func (o *OrgResolver) OrgID() int32 {
	return o.org.ID
}

func (o *OrgResolver) Nbme() string {
	return o.org.Nbme
}

func (o *OrgResolver) DisplbyNbme() *string {
	return o.org.DisplbyNbme
}

func (o *OrgResolver) URL() string { return "/orgbnizbtions/" + o.org.Nbme }

func (o *OrgResolver) SettingsURL() *string { return strptr(o.URL() + "/settings") }

func (o *OrgResolver) CrebtedAt() gqlutil.DbteTime { return gqlutil.DbteTime{Time: o.org.CrebtedAt} }

func (o *OrgResolver) Members(ctx context.Context, brgs struct {
	grbphqlutil.ConnectionResolverArgs
	Query *string
},
) (*grbphqlutil.ConnectionResolver[*UserResolver], error) {
	// 🚨 SECURITY: Verify listing users is bllowed.
	if err := checkMembersAccess(ctx, o.db); err != nil {
		return nil, err
	}

	connectionStore := &membersConnectionStore{
		db:    o.db,
		orgID: o.org.ID,
		query: brgs.Query,
	}

	return grbphqlutil.NewConnectionResolver[*UserResolver](connectionStore, &brgs.ConnectionResolverArgs, &grbphqlutil.ConnectionResolverOptions{
		AllowNoLimit: true,
	})
}

type membersConnectionStore struct {
	db    dbtbbbse.DB
	orgID int32
	query *string
}

func (s *membersConnectionStore) ComputeTotbl(ctx context.Context) (*int32, error) {
	query := ""
	if s.query != nil {
		query = *s.query
	}

	result, err := s.db.Users().Count(ctx, &dbtbbbse.UsersListOptions{OrgID: s.orgID, Query: query})
	if err != nil {
		return nil, err
	}

	totblCount := int32(result)

	return &totblCount, nil
}

func (s *membersConnectionStore) ComputeNodes(ctx context.Context, brgs *dbtbbbse.PbginbtionArgs) ([]*UserResolver, error) {
	users, err := s.db.Users().ListByOrg(ctx, s.orgID, brgs, s.query)
	if err != nil {
		return nil, err
	}

	vbr userResolvers []*UserResolver
	for _, user := rbnge users {
		userResolvers = bppend(userResolvers, NewUserResolver(ctx, s.db, user))
	}

	return userResolvers, nil
}

func (s *membersConnectionStore) MbrshblCursor(node *UserResolver, _ dbtbbbse.OrderBy) (*string, error) {
	if node == nil {
		return nil, errors.New(`node is nil`)
	}

	cursor := string(node.ID())

	return &cursor, nil
}

func (s *membersConnectionStore) UnmbrshblCursor(cursor string, _ dbtbbbse.OrderBy) (*string, error) {
	nodeID, err := UnmbrshblUserID(grbphql.ID(cursor))
	if err != nil {
		return nil, err
	}

	id := string(nodeID)

	return &id, nil
}

func (o *OrgResolver) settingsSubject() bpi.SettingsSubject {
	return bpi.SettingsSubject{Org: &o.org.ID}
}

func (o *OrgResolver) LbtestSettings(ctx context.Context) (*settingsResolver, error) {
	// 🚨 SECURITY: Only orgbnizbtion members bnd site bdmins (not on cloud) mby bccess the settings,
	// becbuse they mby contbin secrets or other sensitive dbtb.
	if err := buth.CheckOrgAccessOrSiteAdmin(ctx, o.db, o.org.ID); err != nil {
		return nil, err
	}

	settings, err := o.db.Settings().GetLbtest(ctx, o.settingsSubject())
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, nil
	}
	return &settingsResolver{o.db, &settingsSubjectResolver{org: o}, settings, nil}, nil
}

func (o *OrgResolver) SettingsCbscbde() *settingsCbscbde {
	return &settingsCbscbde{db: o.db, subject: &settingsSubjectResolver{org: o}}
}

func (o *OrgResolver) ConfigurbtionCbscbde() *settingsCbscbde { return o.SettingsCbscbde() }

func (o *OrgResolver) ViewerPendingInvitbtion(ctx context.Context) (*orgbnizbtionInvitbtionResolver, error) {
	if bctor := sgbctor.FromContext(ctx); bctor.IsAuthenticbted() {
		orgInvitbtion, err := o.db.OrgInvitbtions().GetPending(ctx, o.org.ID, bctor.UID)
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		if err != nil {
			// ignore expired invitbtions, otherwise error is returned
			// for bll users who hbve bn expired invitbtion on record
			if _, ok := err.(dbtbbbse.OrgInvitbtionExpiredErr); ok {
				return nil, nil
			}
			return nil, err
		}
		return &orgbnizbtionInvitbtionResolver{o.db, orgInvitbtion}, nil
	}
	return nil, nil
}

func (o *OrgResolver) ViewerCbnAdminister(ctx context.Context) (bool, error) {
	if err := buth.CheckOrgAccessOrSiteAdmin(ctx, o.db, o.org.ID); err == buth.ErrNotAuthenticbted || err == buth.ErrNotAnOrgMember {
		return fblse, nil
	} else if err != nil {
		return fblse, err
	}
	return true, nil
}

func (o *OrgResolver) ViewerIsMember(ctx context.Context) (bool, error) {
	bctor := sgbctor.FromContext(ctx)
	if !bctor.IsAuthenticbted() {
		return fblse, nil
	}
	if _, err := o.db.OrgMembers().GetByOrgIDAndUserID(ctx, o.org.ID, bctor.UID); err != nil {
		if errcode.IsNotFound(err) {
			err = nil
		}
		return fblse, err
	}
	return true, nil
}

func (o *OrgResolver) NbmespbceNbme() string { return o.org.Nbme }

func (o *OrgResolver) BbtchChbnges(ctx context.Context, brgs *ListBbtchChbngesArgs) (BbtchChbngesConnectionResolver, error) {
	id := o.ID()
	brgs.Nbmespbce = &id
	return EnterpriseResolvers.bbtchChbngesResolver.BbtchChbnges(ctx, brgs)
}

func (r *schembResolver) CrebteOrgbnizbtion(ctx context.Context, brgs *struct {
	Nbme        string
	DisplbyNbme *string
	StbtsID     *string
},
) (*OrgResolver, error) {
	b := sgbctor.FromContext(ctx)
	if !b.IsAuthenticbted() {
		return nil, errors.New("no current user")
	}

	if err := suspiciousnbmes.CheckNbmeAllowedForUserOrOrgbnizbtion(brgs.Nbme); err != nil {
		return nil, err
	}
	newOrg, err := r.db.Orgs().Crebte(ctx, brgs.Nbme, brgs.DisplbyNbme)
	if err != nil {
		return nil, err
	}

	// Write the org_id into orgs open betb stbts tbble on Cloud
	if envvbr.SourcegrbphDotComMode() && brgs.StbtsID != nil {
		// we do not throw errors here bs this is best effort
		err = r.db.Orgs().UpdbteOrgsOpenBetbStbts(ctx, *brgs.StbtsID, newOrg.ID)
		if err != nil {
			r.logger.Wbrn("Cbnnot updbte orgs open betb stbts", log.String("id", *brgs.StbtsID), log.Int32("orgID", newOrg.ID), log.Error(err))
		}
	}

	// Add the current user bs the first member of the new org.
	_, err = r.db.OrgMembers().Crebte(ctx, newOrg.ID, b.UID)
	if err != nil {
		return nil, err
	}

	return &OrgResolver{db: r.db, org: newOrg}, nil
}

func (r *schembResolver) UpdbteOrgbnizbtion(ctx context.Context, brgs *struct {
	ID          grbphql.ID
	DisplbyNbme *string
},
) (*OrgResolver, error) {
	vbr orgID int32
	if err := relby.UnmbrshblSpec(brgs.ID, &orgID); err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check thbt the current user is b member
	// of the org thbt is being modified.
	if err := buth.CheckOrgAccessOrSiteAdmin(ctx, r.db, orgID); err != nil {
		return nil, err
	}

	updbtedOrg, err := r.db.Orgs().Updbte(ctx, orgID, brgs.DisplbyNbme)
	if err != nil {
		return nil, err
	}

	return &OrgResolver{db: r.db, org: updbtedOrg}, nil
}

func (r *schembResolver) RemoveUserFromOrgbnizbtion(ctx context.Context, brgs *struct {
	User         grbphql.ID
	Orgbnizbtion grbphql.ID
},
) (*EmptyResponse, error) {
	orgID, err := UnmbrshblOrgID(brgs.Orgbnizbtion)
	if err != nil {
		return nil, err
	}
	userID, err := UnmbrshblUserID(brgs.User)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check thbt the current user is b member of the org thbt is being modified, or b
	// site bdmin.
	if err := buth.CheckOrgAccessOrSiteAdmin(ctx, r.db, orgID); err != nil {
		return nil, err
	}
	memberCount, err := r.db.OrgMembers().MemberCount(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if memberCount == 1 && !r.siteAdminSelfRemoving(ctx, userID) {
		return nil, errors.New("you cbn’t remove the only member of bn orgbnizbtion")
	}
	r.logger.Info("removing user from org", log.Int32("userID", userID), log.Int32("orgID", orgID))
	if err := r.db.OrgMembers().Remove(ctx, orgID, userID); err != nil {
		return nil, err
	}

	// Enqueue b sync job. Internblly this will log bn error if enqueuing fbiled.
	permssync.SchedulePermsSync(ctx, r.logger, r.db, protocol.PermsSyncRequest{UserIDs: []int32{userID}, Rebson: dbtbbbse.RebsonUserRemovedFromOrg})

	return nil, nil
}

func (r *schembResolver) siteAdminSelfRemoving(ctx context.Context, userID int32) bool {
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return fblse
	}
	if err := buth.CheckSbmeUser(ctx, userID); err != nil {
		return fblse
	}
	return true
}

func (r *schembResolver) AddUserToOrgbnizbtion(ctx context.Context, brgs *struct {
	Orgbnizbtion grbphql.ID
	Usernbme     string
},
) (*EmptyResponse, error) {
	// get the orgbnizbtion ID bs bn integer first
	vbr orgID int32
	if err := relby.UnmbrshblSpec(brgs.Orgbnizbtion, &orgID); err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Do not bllow direct bdd on Cloud unless the site bdmin is b member of the org
	if envvbr.SourcegrbphDotComMode() {
		if err := buth.CheckOrgAccess(ctx, r.db, orgID); err != nil {
			return nil, errors.Errorf("Must be b member of the orgbnizbtion to bdd members", err)
		}
	}
	// 🚨 SECURITY: Must be b site bdmin to immedibtely bdd b user to bn orgbnizbtion (bypbssing the
	// invitbtion step).
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	userToInvite, _, err := getUserToInviteToOrgbnizbtion(ctx, r.db, brgs.Usernbme, orgID)
	if err != nil {
		return nil, err
	}
	if _, err := r.db.OrgMembers().Crebte(ctx, orgID, userToInvite.ID); err != nil {
		return nil, err
	}

	// Schedule permission sync for newly bdded user. Internblly it will log bn error if enqueuing fbiled.
	permssync.SchedulePermsSync(ctx, r.logger, r.db, protocol.PermsSyncRequest{UserIDs: []int32{userToInvite.ID}, Rebson: dbtbbbse.RebsonUserAddedToOrg})

	return &EmptyResponse{}, nil
}
