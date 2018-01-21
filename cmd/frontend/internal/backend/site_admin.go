package backend

import (
	"context"
	"errors"

	store "sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
)

var errMustBeSiteAdmin = errors.New("must be site admin")

// CheckCurrentUserIsSiteAdmin returns an error if the current user is NOT a site admin.
func CheckCurrentUserIsSiteAdmin(ctx context.Context) error {
	user, err := currentUser(ctx)
	if err != nil {
		return err
	}
	// 🚨 SECURITY: Only site admins can make other users site admins (or demote).
	if user == nil || !user.SiteAdmin {
		return errMustBeSiteAdmin
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
	if err := CheckCurrentUserIsSiteAdmin(ctx); err == errMustBeSiteAdmin {
		return errors.New("must be site admin or the self user")
	} else if err != nil {
		return err
	}
	return nil
}

func currentUser(ctx context.Context) (*types.User, error) {
	user, err := store.Users.GetByCurrentAuthUser(ctx)
	if err != nil {
		if _, ok := err.(store.ErrUserNotFound); ok {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}
