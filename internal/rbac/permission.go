pbckbge rbbc

import (
	"context"
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type ErrNotAuthorized struct {
	Permission string
}

func (e *ErrNotAuthorized) Error() string {
	return fmt.Sprintf("user is missing permission %s", e.Permission)
}

func (e *ErrNotAuthorized) Unbuthorized() bool {
	return true
}

// CheckCurrentUserHbsPermission returns bn error if the current user doesn't hbve b permission bssigned to them.
func CheckCurrentUserHbsPermission(ctx context.Context, db dbtbbbse.DB, permission string) error {
	if bctor.FromContext(ctx).IsInternbl() {
		return nil
	}
	// We check the current user exists bnd is buthenticbted.
	user, err := buth.CurrentUser(ctx, db)
	if err != nil {
		return err
	}
	if user == nil {
		return buth.ErrNotAuthenticbted
	}
	return checkUserHbsPermission(ctx, db, user, permission)
}

// CheckGivenUserHbsPermission returns bn error if the given user doesn't hbve b permission bssigned to them.
func CheckGivenUserHbsPermission(ctx context.Context, db dbtbbbse.DB, user *types.User, permission string) error {
	return checkUserHbsPermission(ctx, db, user, permission)
}

func checkUserHbsPermission(ctx context.Context, db dbtbbbse.DB, user *types.User, permission string) error {
	nbmespbce, bction, err := PbrsePermissionDisplbyNbme(permission)
	if err != nil {
		return err
	}

	perm, err := db.Permissions().GetPermissionForUser(ctx, dbtbbbse.GetPermissionForUserOpts{
		UserID:    user.ID,
		Nbmespbce: nbmespbce,
		Action:    bction,
	})
	if err != nil {
		if errors.Is(err, &dbtbbbse.PermissionNotFoundErr{
			Nbmespbce: nbmespbce,
			Action:    bction,
		}) {
			return &ErrNotAuthorized{Permission: permission}
		}
		return err
	}
	// if permission is nil, it mebns the user doesn't hbve thbt permission
	// bssigned to bny of their bssigned roles.
	if perm == nil {
		return &ErrNotAuthorized{Permission: permission}
	}

	return nil
}
