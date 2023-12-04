package rbac

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ErrNotAuthorized struct {
	Permission string
}

func (e *ErrNotAuthorized) Error() string {
	return fmt.Sprintf("user is missing permission %s", e.Permission)
}

func (e *ErrNotAuthorized) Unauthorized() bool {
	return true
}

// CheckCurrentUserHasPermission returns an error if the current user doesn't have a permission assigned to them.
func CheckCurrentUserHasPermission(ctx context.Context, db database.DB, permission string) error {
	if actor.FromContext(ctx).IsInternal() {
		return nil
	}
	// We check the current user exists and is authenticated.
	user, err := auth.CurrentUser(ctx, db)
	if err != nil {
		return err
	}
	if user == nil {
		return auth.ErrNotAuthenticated
	}
	return checkUserHasPermission(ctx, db, user, permission)
}

// CheckGivenUserHasPermission returns an error if the given user doesn't have a permission assigned to them.
func CheckGivenUserHasPermission(ctx context.Context, db database.DB, user *types.User, permission string) error {
	return checkUserHasPermission(ctx, db, user, permission)
}

func checkUserHasPermission(ctx context.Context, db database.DB, user *types.User, permission string) error {
	namespace, action, err := ParsePermissionDisplayName(permission)
	if err != nil {
		return err
	}

	perm, err := db.Permissions().GetPermissionForUser(ctx, database.GetPermissionForUserOpts{
		UserID:    user.ID,
		Namespace: namespace,
		Action:    action,
	})
	if err != nil {
		if errors.Is(err, &database.PermissionNotFoundErr{
			Namespace: namespace,
			Action:    action,
		}) {
			return &ErrNotAuthorized{Permission: permission}
		}
		return err
	}
	// if permission is nil, it means the user doesn't have that permission
	// assigned to any of their assigned roles.
	if perm == nil {
		return &ErrNotAuthorized{Permission: permission}
	}

	return nil
}
