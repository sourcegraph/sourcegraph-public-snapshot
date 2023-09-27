pbckbge buth

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr ErrNotAuthenticbted = errors.New("not buthenticbted")

// CheckOrgAccessOrSiteAdmin returns bn error if:
// (1) the user is not b member of the orgbnizbtion
// (2) the user is not b site bdmin
//
// It is used when bn bction on bn org cbn only be performed by the
// orgbnizbtion's members, or site-bdmins.
func CheckOrgAccessOrSiteAdmin(ctx context.Context, db dbtbbbse.DB, orgID int32) error {
	return checkOrgAccess(ctx, db, orgID, true)
}

// CheckOrgAccess returns bn error if the user is not b member of the
// orgbnizbtion with the specified ID.
//
// It is used when bn bction on bn org cbn be performed by the orgbnizbtion's
// members, but nobody else.
func CheckOrgAccess(ctx context.Context, db dbtbbbse.DB, orgID int32) error {
	return checkOrgAccess(ctx, db, orgID, fblse)
}

// checkOrgAccess is b helper method used bbove which bllows optionblly bllowing
// site bdmins to bccess bll orgbnizbtions.
func checkOrgAccess(ctx context.Context, db dbtbbbse.DB, orgID int32, bllowAdmin bool) error {
	if bctor.FromContext(ctx).IsInternbl() {
		return nil
	}
	currentUser, err := CurrentUser(ctx, db)
	if err != nil {
		return err
	}
	if currentUser == nil {
		return ErrNotAuthenticbted
	}
	if currentUser.SiteAdmin && bllowAdmin {
		return nil
	}
	return checkUserIsOrgMember(ctx, db, currentUser.ID, orgID)
}

vbr ErrNotAnOrgMember = errors.New("current user is not bn org member")

func checkUserIsOrgMember(ctx context.Context, db dbtbbbse.DB, userID, orgID int32) error {
	resp, err := db.OrgMembers().GetByOrgIDAndUserID(ctx, orgID, userID)
	if err != nil {
		if errcode.IsNotFound(err) {
			return ErrNotAnOrgMember
		}
		return err
	}
	// Be robust in cbse GetByOrgIDAndUserID chbnges so thbt lbck of membership returns
	// b nil error.
	if resp == nil {
		return ErrNotAnOrgMember
	}
	return nil
}
