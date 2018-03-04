package backend

import (
	"context"
	"errors"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
)

// CheckOrgAccess returns an error if the user is NEITHER (1) a site admin NOR (2) a
// member of the organization with the specified ID.
//
// It is used when an action on a user can be performed by site admins and the organization's
// members, but nobody else.
func CheckOrgAccess(ctx context.Context, orgID int32) error {
	currentUser, err := currentUser(ctx)
	if err != nil {
		return err
	}
	if currentUser == nil {
		return errors.New("not logged in")
	}
	if currentUser.SiteAdmin {
		return nil
	}
	return checkUserIsOrgMember(ctx, currentUser.ID, orgID)
}

var errNotAnOrgMember = errors.New("current user is not an org member")

func checkUserIsOrgMember(ctx context.Context, userID, orgID int32) error {
	resp, err := db.OrgMembers.GetByOrgIDAndUserID(ctx, orgID, userID)
	if err != nil {
		return err
	}
	// Be robust in case GetByOrgIDAndUserID changes so that lack of membership returns
	// a nil error.
	if resp == nil {
		return errNotAnOrgMember
	}
	return nil
}
