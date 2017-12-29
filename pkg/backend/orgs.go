package backend

import (
	"context"
	"errors"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
)

var Orgs = &orgs{}

type orgs struct{}

func (s *orgs) List(ctx context.Context) (res []*sourcegraph.Org, err error) {
	if Mocks.Orgs.List != nil {
		return Mocks.Orgs.List(ctx)
	}

	// 🚨 SECURITY:  only admins are allowed to use this endpoint
	if err := CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}
	return db.Orgs.List(ctx)
}

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
