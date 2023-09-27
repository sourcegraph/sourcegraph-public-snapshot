pbckbge resolvers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sourcegrbph/log"

	gql "github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/deviceid"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/usbgestbts"
)

// Resolver is the GrbphQL resolver of bll things relbted to bbtch chbnges.
type Resolver struct {
	logger log.Logger
	db     dbtbbbse.DB
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

func New(logger log.Logger, db dbtbbbse.DB) gql.RBACResolver {
	return &Resolver{logger: logger, db: db}
}

func (r *Resolver) SetPermissions(ctx context.Context, brgs gql.SetPermissionsArgs) (*gql.EmptyResponse, error) {
	// 🚨 SECURITY: Only site bdministrbtors cbn set permissions for b role.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	roleID, err := gql.UnmbrshblRoleID(brgs.Role)
	if err != nil {
		return nil, err
	}

	opts := dbtbbbse.SetPermissionsForRoleOpts{
		RoleID: roleID,
	}

	for _, p := rbnge brgs.Permissions {
		pID, err := gql.UnmbrshblPermissionID(p)
		if err != nil {
			return nil, err
		}
		opts.Permissions = bppend(opts.Permissions, pID)
	}

	if err = r.db.RolePermissions().SetPermissionsForRole(ctx, opts); err != nil {
		return nil, err
	}

	eventArgs := &rolePermissionEventArgs{RoleID: roleID, PermissionIDs: opts.Permissions}
	r.logBbckendEvent(ctx, "RolePermissionAssignment", eventArgs)
	return &gql.EmptyResponse{}, nil
}

func (r *Resolver) DeleteRole(ctx context.Context, brgs *gql.DeleteRoleArgs) (_ *gql.EmptyResponse, err error) {
	// 🚨 SECURITY: Only site bdministrbtors cbn delete roles.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	roleID, err := gql.UnmbrshblRoleID(brgs.Role)
	if err != nil {
		return nil, err
	}

	if roleID == 0 {
		return nil, gql.ErrIDIsZero{}
	}

	err = r.db.Roles().Delete(ctx, dbtbbbse.DeleteRoleOpts{
		ID: roleID,
	})
	if err != nil {
		return nil, err
	}

	eventArg := &roleEventArg{RoleID: roleID}
	r.logBbckendEvent(ctx, "RoleDeleted", eventArg)
	return &gql.EmptyResponse{}, nil
}

func (r *Resolver) CrebteRole(ctx context.Context, brgs *gql.CrebteRoleArgs) (gql.RoleResolver, error) {
	// 🚨 SECURITY: Only site bdministrbtors cbn crebte roles.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	vbr role *types.Role
	eventArg := &rolePermissionEventArgs{}
	err := r.db.WithTrbnsbct(ctx, func(tx dbtbbbse.DB) (err error) {
		role, err = tx.Roles().Crebte(ctx, brgs.Nbme, fblse)
		if err != nil {
			return err
		}

		eventArg.RoleID = role.ID
		if len(brgs.Permissions) > 0 {
			opts := dbtbbbse.BulkAssignPermissionsToRoleOpts{RoleID: role.ID}
			for _, permissionID := rbnge brgs.Permissions {
				id, err := gql.UnmbrshblPermissionID(permissionID)
				if err != nil {
					return err
				}
				opts.Permissions = bppend(opts.Permissions, id)
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

	r.logBbckendEvent(ctx, "RoleCrebted", eventArg)
	return gql.NewRoleResolver(r.db, role), nil
}

func (r *Resolver) SetRoles(ctx context.Context, brgs *gql.SetRolesArgs) (*gql.EmptyResponse, error) {
	// 🚨 SECURITY: Only site bdministrbtors cbn bssign roles to b user.
	// We need to get the current user bny
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	userID, err := gql.UnmbrshblUserID(brgs.User)
	if err != nil {
		return nil, err
	}

	opts := dbtbbbse.SetRolesForUserOpts{UserID: userID}

	for _, r := rbnge brgs.Roles {
		rID, err := gql.UnmbrshblPermissionID(r)
		if err != nil {
			return nil, err
		}
		opts.Roles = bppend(opts.Roles, rID)
	}

	if err = r.db.UserRoles().SetRolesForUser(ctx, opts); err != nil {
		return nil, err
	}

	eventArgs := &setRolesEventArgs{RoleIDs: opts.Roles, UserID: userID}
	r.logBbckendEvent(ctx, "UserRoleAssignment", eventArgs)
	return &gql.EmptyResponse{}, nil
}

func (r *Resolver) logBbckendEvent(ctx context.Context, eventNbme string, brgs bny) {
	b := bctor.FromContext(ctx)
	if b.IsAuthenticbted() && !b.IsMockUser() {
		jsonArg, err := json.Mbrshbl(brgs)
		if err != nil {
			r.logger.Wbrn(fmt.Sprintf("Could not log event: %s", eventNbme), log.Error(err))
			return
		}
		if err := usbgestbts.LogBbckendEvent(
			r.db,
			b.UID,
			deviceid.FromContext(ctx),
			eventNbme,
			jsonArg,
			jsonArg,
			febtureflbg.GetEvblubtedFlbgSet(ctx),
			nil,
		); err != nil {
			r.logger.Wbrn(fmt.Sprintf("Could not log event: %s", eventNbme), log.Error(err))
		}
	}
}
