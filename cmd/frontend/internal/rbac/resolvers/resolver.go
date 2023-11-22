package resolvers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/log"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/deviceid"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/usagestats"
)

// Resolver is the GraphQL resolver of all things related to batch changes.
type Resolver struct {
	logger log.Logger
	db     database.DB
}

type roleEventArg struct {
	RoleID int32 `json:"role_id"`
}

type rolePermissionEventArgs struct {
	RoleID        int32   `json:"role_id"`
	PermissionIDs []int32 `json:"permission_ids"`
}

type setRolesEventArgs struct {
	UserID  int32   `json:"user_id"`
	RoleIDs []int32 `json:"role_ids"`
}

func New(logger log.Logger, db database.DB) gql.RBACResolver {
	return &Resolver{logger: logger, db: db}
}

func (r *Resolver) SetPermissions(ctx context.Context, args gql.SetPermissionsArgs) (*gql.EmptyResponse, error) {
	// 🚨 SECURITY: Only site administrators can set permissions for a role.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	roleID, err := gql.UnmarshalRoleID(args.Role)
	if err != nil {
		return nil, err
	}

	opts := database.SetPermissionsForRoleOpts{
		RoleID: roleID,
	}

	for _, p := range args.Permissions {
		pID, err := gql.UnmarshalPermissionID(p)
		if err != nil {
			return nil, err
		}
		opts.Permissions = append(opts.Permissions, pID)
	}

	if err = r.db.RolePermissions().SetPermissionsForRole(ctx, opts); err != nil {
		return nil, err
	}

	eventArgs := &rolePermissionEventArgs{RoleID: roleID, PermissionIDs: opts.Permissions}
	r.logBackendEvent(ctx, "RolePermissionAssignment", eventArgs)
	return &gql.EmptyResponse{}, nil
}

func (r *Resolver) DeleteRole(ctx context.Context, args *gql.DeleteRoleArgs) (_ *gql.EmptyResponse, err error) {
	// 🚨 SECURITY: Only site administrators can delete roles.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	roleID, err := gql.UnmarshalRoleID(args.Role)
	if err != nil {
		return nil, err
	}

	if roleID == 0 {
		return nil, gql.ErrIDIsZero{}
	}

	err = r.db.Roles().Delete(ctx, database.DeleteRoleOpts{
		ID: roleID,
	})
	if err != nil {
		return nil, err
	}

	eventArg := &roleEventArg{RoleID: roleID}
	r.logBackendEvent(ctx, "RoleDeleted", eventArg)
	return &gql.EmptyResponse{}, nil
}

func (r *Resolver) CreateRole(ctx context.Context, args *gql.CreateRoleArgs) (gql.RoleResolver, error) {
	// 🚨 SECURITY: Only site administrators can create roles.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	var role *types.Role
	eventArg := &rolePermissionEventArgs{}
	err := r.db.WithTransact(ctx, func(tx database.DB) (err error) {
		role, err = tx.Roles().Create(ctx, args.Name, false)
		if err != nil {
			return err
		}

		eventArg.RoleID = role.ID
		if len(args.Permissions) > 0 {
			opts := database.BulkAssignPermissionsToRoleOpts{RoleID: role.ID}
			for _, permissionID := range args.Permissions {
				id, err := gql.UnmarshalPermissionID(permissionID)
				if err != nil {
					return err
				}
				opts.Permissions = append(opts.Permissions, id)
			}
			eventArg.PermissionIDs = opts.Permissions
			err = tx.RolePermissions().BulkAssignPermissionsToRole(ctx, opts)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	r.logBackendEvent(ctx, "RoleCreated", eventArg)
	return gql.NewRoleResolver(r.db, role), nil
}

func (r *Resolver) SetRoles(ctx context.Context, args *gql.SetRolesArgs) (*gql.EmptyResponse, error) {
	// 🚨 SECURITY: Only site administrators can assign roles to a user.
	// We need to get the current user any
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	userID, err := gql.UnmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}

	opts := database.SetRolesForUserOpts{UserID: userID}

	for _, r := range args.Roles {
		rID, err := gql.UnmarshalPermissionID(r)
		if err != nil {
			return nil, err
		}
		opts.Roles = append(opts.Roles, rID)
	}

	if err = r.db.UserRoles().SetRolesForUser(ctx, opts); err != nil {
		return nil, err
	}

	eventArgs := &setRolesEventArgs{RoleIDs: opts.Roles, UserID: userID}
	r.logBackendEvent(ctx, "UserRoleAssignment", eventArgs)
	return &gql.EmptyResponse{}, nil
}

func (r *Resolver) logBackendEvent(ctx context.Context, eventName string, args any) {
	a := actor.FromContext(ctx)
	if a.IsAuthenticated() && !a.IsMockUser() {
		jsonArg, err := json.Marshal(args)
		if err != nil {
			r.logger.Warn(fmt.Sprintf("Could not log event: %s", eventName), log.Error(err))
			return
		}
		if err := usagestats.LogBackendEvent(
			r.db,
			a.UID,
			deviceid.FromContext(ctx),
			eventName,
			jsonArg,
			jsonArg,
			featureflag.GetEvaluatedFlagSet(ctx),
			nil,
		); err != nil {
			r.logger.Warn(fmt.Sprintf("Could not log event: %s", eventName), log.Error(err))
		}
	}
}
