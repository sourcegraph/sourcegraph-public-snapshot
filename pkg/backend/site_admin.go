package backend

import (
	"context"
	"errors"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

// CheckCurrentUserIsSiteAdmin returns an error if the current user is NOT a site admin.
func CheckCurrentUserIsSiteAdmin(ctx context.Context) error {
	user, err := currentUser(ctx)
	if err != nil {
		return err
	}
	// 🚨 SECURITY: Only site admins can make other users site admins (or demote).
	if user == nil || !user.SiteAdmin {
		return errors.New("must be site admin")
	}
	return nil
}

func currentUser(ctx context.Context) (*sourcegraph.User, error) {
	user, err := store.Users.GetByCurrentAuthUser(ctx)
	if err != nil {
		if _, ok := err.(store.ErrUserNotFound); ok {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}
