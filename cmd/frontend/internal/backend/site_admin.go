package backend

import (
	"context"
	"errors"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
)

var ErrMustBeSiteAdmin = errors.New("must be site admin")

// CheckCurrentUserIsSiteAdmin returns an error if the current user is NOT a site admin.
func CheckCurrentUserIsSiteAdmin(ctx context.Context) error {
	user, err := currentUser(ctx)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrNotAuthenticated
	}
	if !user.SiteAdmin {
		return ErrMustBeSiteAdmin
	}
	return nil
}

// CheckSiteAdminOrSameUser returns an error if the user is NEITHER (1) a
// site admin NOR (2) the user specified by subjectUserID.
//
// It is used when an action on a user can be performed by site admins and the
// user themselves, but nobody else.
func CheckSiteAdminOrSameUser(ctx context.Context, subjectUserID int32) error {
	actor := actor.FromContext(ctx)
	if actor.IsAuthenticated() && actor.UID == subjectUserID {
		return nil
	}
	return CheckCurrentUserIsSiteAdmin(ctx)
}

func currentUser(ctx context.Context) (*types.User, error) {
	user, err := db.Users.GetByCurrentAuthUser(ctx)
	if err != nil {
		if errcode.IsNotFound(err) || err == db.ErrNoCurrentUser {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}
