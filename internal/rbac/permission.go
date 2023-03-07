package rbac

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var ErrNotAuthorized = errors.New("not authorized")

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

	namespace, action, err := parsePermissionDisplayName(permission)
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
			return ErrNotAuthorized
		}
		return err
	}
	// if permission is nil, it means the user doesn't have that permission
	// assigned to any of their assigned roles.
	if perm == nil {
		return ErrNotAuthorized
	}

	return nil
}
