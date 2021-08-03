package backend

import (
	"context"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
)

var ErrNotAuthenticated = errors.New("not authenticated")

// CheckOrgAccessOrSiteAdmin returns an error if the user is NEITHER (1) a site
// admin NOR (2) a member of the organization with the specified ID.
//
// It is used when an action on a user can be performed by site admins and the
// organization's members, but nobody else.
func CheckOrgAccessOrSiteAdmin(ctx context.Context, db dbutil.DB, orgID int32) error {
	return checkOrgAccess(ctx, db, orgID, true)
}

// CheckOrgAccess returns an error if the user is not a member of the
// organization with the specified ID.
//
// It is used when an action on a user can be performed by the organization's
// members, but nobody else.
func CheckOrgAccess(ctx context.Context, db dbutil.DB, orgID int32) error {
	return checkOrgAccess(ctx, db, orgID, false)
}

// checkOrgAccess is a helper method used above which allows optionally allowing
// site admins to access all organisations.
func checkOrgAccess(ctx context.Context, db dbutil.DB, orgID int32, allowAdmin bool) error {
	if actor.FromContext(ctx).IsInternal() {
		return nil
	}
	currentUser, err := CurrentUser(ctx, db)
	if err != nil {
		return err
	}
	if currentUser == nil {
		return ErrNotAuthenticated
	}
	if currentUser.SiteAdmin && allowAdmin {
		return nil
	}
	return checkUserIsOrgMember(ctx, db, currentUser.ID, orgID)
}

var ErrNotAnOrgMember = errors.New("current user is not an org member")

func checkUserIsOrgMember(ctx context.Context, db dbutil.DB, userID, orgID int32) error {
	resp, err := database.OrgMembers(db).GetByOrgIDAndUserID(ctx, orgID, userID)
	if err != nil {
		if errcode.IsNotFound(err) {
			return ErrNotAnOrgMember
		}
		return err
	}
	// Be robust in case GetByOrgIDAndUserID changes so that lack of membership returns
	// a nil error.
	if resp == nil {
		return ErrNotAnOrgMember
	}
	return nil
}
