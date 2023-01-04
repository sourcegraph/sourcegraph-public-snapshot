package rbac

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func AssignRolesForUser(ctx context.Context, db database.DB, user *types.User) error {

	// We fetch all read only roles (DEFAULT and SITE_ADMINISTRATOR)
	roles, err := db.Roles().List(ctx, database.RolesListOptions{
		ReadOnly: true,
		LimitOffset: &database.LimitOffset{
			// We add a limit of 2 because this query is ordered by the created_at and we'll only
			// have two readonly roles for now. In the event that new readonly roles get added in the future,
			// this query will still return the correct response since the roles will be ordered by
			// created_at and the DEFAULT and SITE_ADMINISTRATOR roles will always be created before any other roles.
			Limit: 2,
		},
	})
	if err != nil {
		return err
	}

	var roleIDs []int32

	for _, role := range roles {
		if user.SiteAdmin && role.Name == database.SiteAdministratorRole {
			roleIDs = append(roleIDs, role.ID)
		}

		if role.Name == database.DefaultRole {
			roleIDs = append(roleIDs, role.ID)
		}
	}

	_, err = db.UserRoles().CreateMultipleUserRolesForUser(ctx, database.CreateMultipleUserRolesForUserOpts{
		UserID:  user.ID,
		RoleIDs: roleIDs,
	})
	return err
}

func HandleUserSiteAdminStatus(ctx context.Context, db database.DB, isSiteAdmin bool, userID int32) error {
	adminRole, err := db.Roles().Get(ctx, database.GetRoleOpts{
		Name: database.SiteAdministratorRole,
	})
	if err != nil {
		return err
	}

	if isSiteAdmin {
		// We add the SITE_ADMINISTRATOR role for this user
		_, err = db.UserRoles().Create(ctx, database.CreateUserRoleOpts{
			UserID: userID,
			RoleID: adminRole.ID,
		})
		if err != nil {
			return err
		}
	} else {
		// We remove the site administrator role for this user
		err = db.UserRoles().Delete(ctx, database.DeleteUserRoleOpts{
			UserID: userID,
			RoleID: adminRole.ID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
