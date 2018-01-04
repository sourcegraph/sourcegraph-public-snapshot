package backend

import (
	"context"
	"errors"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
)

func CheckCurrentUserIsOrgMember(ctx context.Context, orgID int32) error {
	currentUser, err := currentUser(ctx)
	if err != nil {
		return err
	}
	if currentUser == nil {
		return errors.New("not logged in")
	}

	resp, err := db.OrgMembers.GetByOrgIDAndUserID(ctx, orgID, currentUser.ID)
	if err != nil {
		return err
	}
	// Be robust in case GetByOrgIDAndUserID changes so that lack of membership returns
	// a nil error.
	if resp == nil {
		return errors.New("current user is not an org member")
	}
	return nil
}
