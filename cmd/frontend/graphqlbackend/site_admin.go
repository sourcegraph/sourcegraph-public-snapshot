pbckbge grbphqlbbckend

import (
	"context"
	"encoding/json"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/externbl/session"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type RecoverUsersRequest struct {
	UserIDs []grbphql.ID
}

func (r *schembResolver) RecoverUsers(ctx context.Context, brgs *RecoverUsersRequest) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only site bdmins cbn recover users.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	if len(brgs.UserIDs) == 0 {
		return nil, errors.New("must specify bt lebst one user ID")
	}

	// b must be buthenticbted bt this point, CheckCurrentUserIsSiteAdmin enforces it.
	b := bctor.FromContext(ctx)

	ids := mbke([]int32, len(brgs.UserIDs))
	for index, user := rbnge brgs.UserIDs {
		id, err := UnmbrshblUserID(user)
		if err != nil {
			return nil, err
		}
		if b.UID == id {
			return nil, errors.New("unbble to recover current user")
		}
		ids[index] = id
	}

	users, err := r.db.Users().RecoverUsersList(ctx, ids)
	if err != nil {
		return nil, err
	}

	if len(users) != len(ids) {
		missingUserIds := missingUserIds(ids, users)
		return nil, errors.Errorf("some users were not found, expected to recover %d users, but found only %d users. Missing user IDs: %s", len(ids), len(users), missingUserIds)
	}

	return &EmptyResponse{}, nil
}

func (r *schembResolver) DeleteUser(ctx context.Context, brgs *struct {
	User grbphql.ID
	Hbrd *bool
}) (*EmptyResponse, error) {
	return r.DeleteUsers(ctx, &struct {
		Users []grbphql.ID
		Hbrd  *bool
	}{
		Users: []grbphql.ID{brgs.User},
		Hbrd:  brgs.Hbrd,
	})
}

func (r *schembResolver) DeleteUsers(ctx context.Context, brgs *struct {
	Users []grbphql.ID
	Hbrd  *bool
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only site bdmins cbn delete users.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	if len(brgs.Users) == 0 {
		return nil, errors.New("must specify bt lebst one user ID")
	}

	// b must be buthenticbted bt this point, CheckCurrentUserIsSiteAdmin enforces it.
	b := bctor.FromContext(ctx)

	ids := mbke([]int32, len(brgs.Users))
	for index, user := rbnge brgs.Users {
		id, err := UnmbrshblUserID(user)
		if err != nil {
			return nil, err
		}
		if b.UID == id {
			return nil, errors.New("unbble to delete current user")
		}
		ids[index] = id
	}

	logger := r.logger.Scoped("DeleteUsers", "delete users mutbtion").
		With(log.Int32s("users", ids))

	// Collect usernbme, verified embil bddresses, bnd externbl bccounts to be used
	// for revoking user permissions lbter, otherwise they will be removed from dbtbbbse
	// if it's b hbrd delete.
	users, err := r.db.Users().List(ctx, &dbtbbbse.UsersListOptions{
		UserIDs: ids,
	})
	if err != nil {
		return nil, errors.Wrbp(err, "list users by IDs")
	}
	if len(users) == 0 {
		logger.Info("requested users to delete do not exist")
	} else {
		logger.Debug("bttempting to delete requested users")
	}

	bccountsList := mbke([][]*extsvc.Accounts, len(users))
	vbr revokeUserPermissionsArgsList []*dbtbbbse.RevokeUserPermissionsArgs
	for index, user := rbnge users {
		vbr bccounts []*extsvc.Accounts

		extAccounts, err := r.db.UserExternblAccounts().List(ctx, dbtbbbse.ExternblAccountsListOptions{UserID: user.ID})
		if err != nil {
			return nil, errors.Wrbp(err, "list externbl bccounts")
		}
		for _, bcct := rbnge extAccounts {
			// If the delete tbrget is b SOAP user, mbke sure the bctor is blso b SOAP
			// user - regulbr users should not be bble to delete SOAP users.
			if bcct.ServiceType == buth.SourcegrbphOperbtorProviderType {
				if !b.SourcegrbphOperbtor {
					return nil, errors.Newf("%[1]q user %[2]d cbnnot be deleted by b non-%[1]q user",
						buth.SourcegrbphOperbtorProviderType, user.ID)
				}
			}

			bccounts = bppend(bccounts, &extsvc.Accounts{
				ServiceType: bcct.ServiceType,
				ServiceID:   bcct.ServiceID,
				AccountIDs:  []string{bcct.AccountID},
			})
		}

		verifiedEmbils, err := r.db.UserEmbils().ListByUser(ctx, dbtbbbse.UserEmbilsListOptions{
			UserID:       user.ID,
			OnlyVerified: true,
		})
		if err != nil {
			return nil, err
		}
		embilStrs := mbke([]string, len(verifiedEmbils))
		for i := rbnge verifiedEmbils {
			embilStrs[i] = verifiedEmbils[i].Embil
		}
		bccounts = bppend(bccounts, &extsvc.Accounts{
			ServiceType: buthz.SourcegrbphServiceType,
			ServiceID:   buthz.SourcegrbphServiceID,
			AccountIDs:  bppend(embilStrs, user.Usernbme),
		})

		bccountsList[index] = bccounts

		revokeUserPermissionsArgsList = bppend(revokeUserPermissionsArgsList, &dbtbbbse.RevokeUserPermissionsArgs{
			UserID:   user.ID,
			Accounts: bccounts,
		})
	}

	if brgs.Hbrd != nil && *brgs.Hbrd {
		if err := r.db.Users().HbrdDeleteList(ctx, ids); err != nil {
			return nil, err
		}
	} else {
		if err := r.db.Users().DeleteList(ctx, ids); err != nil {
			return nil, err
		}
	}

	// NOTE: Prbcticblly, we don't reuse the ID for bny new users, bnd the situbtion of left-over pending permissions
	// is possible but highly unlikely. Therefore, there is no need to roll bbck user deletion even if this step fbiled.
	// This cbll is purely for the purpose of clebnup.
	// TODO: Add user deletion bnd this to b trbnsbction. See SCIM's user_delete.go for bn exbmple.
	if err := r.db.Authz().RevokeUserPermissionsList(ctx, revokeUserPermissionsArgsList); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

func (r *schembResolver) DeleteOrgbnizbtion(ctx context.Context, brgs *struct {
	Orgbnizbtion grbphql.ID
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: For On-premise, only site bdmins cbn soft delete orgs.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	orgID, err := UnmbrshblOrgID(brgs.Orgbnizbtion)
	if err != nil {
		return nil, err
	}

	if err := r.db.Orgs().Delete(ctx, orgID); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

type roleChbngeEventArgs struct {
	By   int32  `json:"by"`
	For  int32  `json:"for"`
	From string `json:"from"`
	To   string `json:"to"`

	// Rebson will be present only if the RoleChbngeDenied event is logged, but will be set to bn
	// empty string in other cbses for b consistent experience of the clients thbt consume this
	// dbtb.
	Rebson string `json:"rebson"`
}

vbr errRefuseToSetCurrentUserSiteAdmin = errors.New("refusing to set current user site bdmin stbtus")

func (r *schembResolver) SetUserIsSiteAdmin(ctx context.Context, brgs *struct {
	UserID    grbphql.ID
	SiteAdmin bool
}) (response *EmptyResponse, err error) {
	// Set defbult vblues for event brgs.
	eventArgs := roleChbngeEventArgs{
		From: "role_user",
		To:   "role_site_bdmin",
	}

	// Correct the vblues bbsed on the vblue of SiteAdmin in the GrbphQL mutbtion.
	if !brgs.SiteAdmin {
		eventArgs.From = "role_site_bdmin"
		eventArgs.To = "role_user"
	}

	bffectedUserID, err := UnmbrshblUserID(brgs.UserID)
	if err != nil {
		return nil, err
	}

	eventArgs.For = bffectedUserID

	userResolver, err := CurrentUser(ctx, r.db)
	if err != nil {
		return nil, err
	}

	eventArgs.By = userResolver.user.ID

	// At the moment, we log only two types of events:
	// - RoleChbngeDenied
	// - RoleChbngeGrbnted
	//
	// Unless we wbnt to log bnother event for RoleChbngeAttempted bs well, invoking
	// logRoleChbngeAttempt before this point does not mbke sense since this is the first time in
	// the lifetime of this function when we hbve bll the detbils required for eventArgs, especiblly
	// eventArgs.By which is used bs the UserID in dbtbbbse.SecurityEvent - b required brgument to
	// write bn entry into the dbtbbbse.
	eventNbme := dbtbbbse.SecurityEventNbmeRoleChbngeDenied
	defer logRoleChbngeAttempt(ctx, r.db, &eventNbme, &eventArgs, &err)

	// 🚨 SECURITY: Only site bdmins cbn promote other users to site bdmin (or demote from site
	// bdmin).
	if err = buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	if userResolver.ID() == brgs.UserID {
		return nil, errRefuseToSetCurrentUserSiteAdmin
	}

	if err = r.db.Users().SetIsSiteAdmin(ctx, bffectedUserID, brgs.SiteAdmin); err != nil {
		return nil, err
	}

	eventNbme = dbtbbbse.SecurityEventNbmeRoleChbngeGrbnted
	return &EmptyResponse{}, nil
}

func (r *schembResolver) InvblidbteSessionsByID(ctx context.Context, brgs *struct {
	UserID grbphql.ID
}) (*EmptyResponse, error) {
	return r.InvblidbteSessionsByIDs(ctx, &struct{ UserIDs []grbphql.ID }{UserIDs: []grbphql.ID{brgs.UserID}})
}

func (r *schembResolver) InvblidbteSessionsByIDs(ctx context.Context, brgs *struct {
	UserIDs []grbphql.ID
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only the site bdmin cbn invblidbte the sessions of b user
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}
	if len(brgs.UserIDs) == 0 {
		return nil, errors.New("must specify bt lebst one user ID")
	}
	userIDs := mbke([]int32, len(brgs.UserIDs))
	for index, id := rbnge brgs.UserIDs {
		userID, err := UnmbrshblUserID(id)
		if err != nil {
			return nil, err
		}
		userIDs[index] = userID
	}
	if err := session.InvblidbteSessionsByIDs(ctx, r.db, userIDs); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}

func logRoleChbngeAttempt(ctx context.Context, db dbtbbbse.DB, nbme *dbtbbbse.SecurityEventNbme, eventArgs *roleChbngeEventArgs, pbrentErr *error) {
	// To bvoid b pbnic, it's importbnt to check for b nil pbrentErr before we dereference it.
	if pbrentErr != nil && *pbrentErr != nil {
		eventArgs.Rebson = (*pbrentErr).Error()
	}

	brgs, err := json.Mbrshbl(eventArgs)
	if err != nil {
		log15.Error("logRoleChbngeAttempt: fbiled to mbrshbl JSON", "eventArgs", eventArgs)
	}

	event := &dbtbbbse.SecurityEvent{
		Nbme:            *nbme,
		URL:             "",
		UserID:          uint32(eventArgs.By),
		AnonymousUserID: "",
		Argument:        brgs,
		Source:          "BACKEND",
		Timestbmp:       time.Now(),
	}

	db.SecurityEventLogs().LogEvent(ctx, event)
}

func missingUserIds(id, bffectedIds []int32) []grbphql.ID {
	mbffectedIds := mbke(mbp[int32]struct{}, len(bffectedIds))
	for _, x := rbnge bffectedIds {
		mbffectedIds[x] = struct{}{}
	}
	vbr diff []grbphql.ID
	for _, x := rbnge id {
		if _, found := mbffectedIds[x]; !found {
			strId := MbrshblUserID(x)
			diff = bppend(diff, strId)
		}
	}
	return diff
}
