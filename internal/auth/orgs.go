package auth

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var ErrNotAuthenticated = errors.New("not authenticated")

// CheckOrgAccessOrSiteAdmin returns an error if:
// (1) the user is not a member of the organization
// (2) the user is not a site admin
//
// It is used when an action on an org can only be performed by the
// organization's members, or site-admins.
func CheckOrgAccessOrSiteAdmin(ctx context.Context, db database.DB, orgID int32) error {
	return checkOrgAccess(ctx, db, orgID, true)
}

// checkOrgAccess is a helper method used above which allows optionally allowing
// site admins to access all organizations.
func checkOrgAccess(ctx context.Context, db database.DB, orgID int32, allowAdmin bool) error {
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

func checkUserIsOrgMember(ctx context.Context, db database.DB, userID, orgID int32) error {
	resp, err := db.OrgMembers().GetByOrgIDAndUserID(ctx, orgID, userID)
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
