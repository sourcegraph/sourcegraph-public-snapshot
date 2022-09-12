package backend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var ErrNotAuthenticated = errors.New("not authenticated")

// CheckOrgAccessOrSiteAdmin returns an error if:
// (1) if we are on Cloud instance and the user is not a member of the organization
// (2) if we are NOT on Cloud and
//
//	(a) the user is not a member of the organization
//	(b) the user is not a site admin
//
// It is used when an action on an org can only be performed by the
// organization's members, (or site-admins - not on Cloud).
func CheckOrgAccessOrSiteAdmin(ctx context.Context, db database.DB, orgID int32) error {
	allowAdmin := !envvar.SourcegraphDotComMode()
	return checkOrgAccess(ctx, db, orgID, allowAdmin)
}

// CheckOrgAccess returns an error if the user is not a member of the
// organization with the specified ID.
//
// It is used when an action on an org can be performed by the organization's
// members, but nobody else.
func CheckOrgAccess(ctx context.Context, db database.DB, orgID int32) error {
	return checkOrgAccess(ctx, db, orgID, false)
}

// checkOrgAccess is a helper method used above which allows optionally allowing
// site admins to access all organisations.
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
