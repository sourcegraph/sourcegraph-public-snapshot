package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var rolePermissionInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("role_id"),
	sqlf.Sprintf("permission_id"),
}

var rolePermissionColumns = []*sqlf.Query{
	sqlf.Sprintf("role_permissions.role_id"),
	sqlf.Sprintf("role_permissions.permission_id"),
	sqlf.Sprintf("role_permissions.created_at"),
}

type RolePermissionOpts struct {
	PermissionID int32
	RoleID       int32
}

type (
	AssignRolePermissionOpts RolePermissionOpts
	RevokeRolePermissionOpts RolePermissionOpts
	GetRolePermissionOpts    RolePermissionOpts
)

type AssignToSystemRoleOpts struct {
	Role         types.SystemRole
	PermissionID int32
}

type BulkAssignPermissionsToSystemRolesOpts struct {
	Roles        []types.SystemRole
	PermissionID int32
}

type BulkOperationOpts struct {
	RoleID      int32
	Permissions []int32
}

type (
	BulkAssignPermissionsToRoleOpts  BulkOperationOpts
	BulkRevokePermissionsForRoleOpts BulkOperationOpts
	SetPermissionsForRoleOpts        BulkOperationOpts
)

type RolePermissionStore interface {
	basestore.ShareableStore

	// Assign is used to assign a permission to a role.
	Assign(ctx context.Context, opts AssignRolePermissionOpts) error
	// AssignToSystemRole is used to assign a permission to a system role.
	AssignToSystemRole(ctx context.Context, opts AssignToSystemRoleOpts) error
	// BulkAssignPermissionsToRole is used to assign multiple permissions to a role.
	BulkAssignPermissionsToRole(ctx context.Context, opts BulkAssignPermissionsToRoleOpts) error
	// BulkAssignPermissionsToSystemRoles is used to assign a permission to multiple system roles.
	BulkAssignPermissionsToSystemRoles(ctx context.Context, opts BulkAssignPermissionsToSystemRolesOpts) error
	// BulkRevokePermissionsForRole revokes bulk permissions assigned to a role.
	BulkRevokePermissionsForRole(ctx context.Context, opts BulkRevokePermissionsForRoleOpts) error
	// GetByRoleIDAndPermissionID returns one RolePermission associated with the provided role and permission.
	GetByRoleIDAndPermissionID(ctx context.Context, opts GetRolePermissionOpts) (*types.RolePermission, error)
	// GetByRoleID returns all RolePermission associated with the provided role ID
	GetByRoleID(ctx context.Context, opts GetRolePermissionOpts) ([]*types.RolePermission, error)
	// GetByPermissionID returns all RolePermission associated with the provided permission ID
	GetByPermissionID(ctx context.Context, opts GetRolePermissionOpts) ([]*types.RolePermission, error)
	// Revoke deletes the permission and role relationship from the database.
	Revoke(ctx context.Context, opts RevokeRolePermissionOpts) error
	// SetPermissionsForRole is used to sync all permissions assigned to a role. It removes any permission not
	// included in the options and assigns permissions that aren't yet assigned in the database.
	SetPermissionsForRole(ctx context.Context, opts SetPermissionsForRoleOpts) error
	// WithTransact creates a transaction for the RolePermissionStore.
	WithTransact(context.Context, func(RolePermissionStore) error) error
	// With is used to merge the store with another to pull data via other stores.
	With(basestore.ShareableStore) RolePermissionStore
}

type rolePermissionStore struct {
	*basestore.Store
}

var _ RolePermissionStore = &rolePermissionStore{}

func RolePermissionsWith(other basestore.ShareableStore) RolePermissionStore {
	return &rolePermissionStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (rp *rolePermissionStore) With(other basestore.ShareableStore) RolePermissionStore {
	return &rolePermissionStore{Store: rp.Store.With(other)}
}

func (rp *rolePermissionStore) WithTransact(ctx context.Context, f func(RolePermissionStore) error) error {
	return rp.Store.WithTransact(ctx, func(tx *basestore.Store) error {
		return f(&rolePermissionStore{Store: tx})
	})
}

const deleteRolePermissionQueryFmtStr = `
DELETE FROM role_permissions
WHERE %s
`

func (rp *rolePermissionStore) Revoke(ctx context.Context, opts RevokeRolePermissionOpts) error {
	if opts.PermissionID == 0 {
		return errors.New("missing permission id")
	}

	// First we check that we're not modifying the site administrator role, which
	// should not be modified except by the `UpdatePermissions` startup process.
	err := checkRoleExistsAndIsNotSiteAdminRole(ctx, rp, opts.RoleID)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(
		deleteRolePermissionQueryFmtStr,
		sqlf.Sprintf("role_permissions.permission_id = %s AND role_permissions.role_id = %s", opts.PermissionID, opts.RoleID),
	)

	result, err := rp.ExecResult(ctx, q)
	if err != nil {
		return errors.Wrap(err, "running delete query")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "checking deleted rows")
	}

	if rowsAffected == 0 {
		return errors.Wrap(&RolePermissionNotFoundErr{opts.PermissionID, opts.RoleID}, "failed to revoke role permission")
	}

	return nil
}

func (rp *rolePermissionStore) BulkRevokePermissionsForRole(ctx context.Context, opts BulkRevokePermissionsForRoleOpts) error {
	// First we check that we're not modifying the site administrator role, which
	// should not be modified except by the `UpdatePermissions` startup process.
	err := checkRoleExistsAndIsNotSiteAdminRole(ctx, rp, opts.RoleID)
	if err != nil {
		return err
	}

	if len(opts.Permissions) == 0 {
		return errors.New("missing permissions")
	}

	var preds []*sqlf.Query

	var permissionIDs []*sqlf.Query
	for _, permission := range opts.Permissions {
		permissionIDs = append(permissionIDs, sqlf.Sprintf("%s", permission))
	}

	preds = append(preds, sqlf.Sprintf("role_id = %s", opts.RoleID))
	preds = append(preds, sqlf.Sprintf("permission_id IN ( %s )", sqlf.Join(permissionIDs, ", ")))

	q := sqlf.Sprintf(
		deleteRolePermissionQueryFmtStr,
		sqlf.Join(preds, " AND "),
	)

	return rp.Exec(ctx, q)
}

func scanRolePermission(sc dbutil.Scanner) (*types.RolePermission, error) {
	var rp types.RolePermission
	if err := sc.Scan(
		&rp.RoleID,
		&rp.PermissionID,
		&rp.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &rp, nil
}

const getRolePermissionQueryFmtStr = `
SELECT
	%s
FROM role_permissions
WHERE %s
`

func (rp *rolePermissionStore) get(ctx context.Context, w *sqlf.Query) ([]*types.RolePermission, error) {
	q := sqlf.Sprintf(
		getRolePermissionQueryFmtStr,
		sqlf.Join(rolePermissionColumns, ", "),
		w,
	)

	var scanRolePermissions = basestore.NewSliceScanner(scanRolePermission)
	return scanRolePermissions(rp.Query(ctx, q))
}

func (rp *rolePermissionStore) GetByPermissionID(ctx context.Context, opts GetRolePermissionOpts) ([]*types.RolePermission, error) {
	if opts.PermissionID == 0 {
		return nil, errors.New("missing permission id")
	}

	return rp.get(ctx, sqlf.Sprintf("permission_id = %s", opts.PermissionID))
}

func (rp *rolePermissionStore) GetByRoleID(ctx context.Context, opts GetRolePermissionOpts) ([]*types.RolePermission, error) {
	if opts.RoleID == 0 {
		return nil, errors.New("missing role id")
	}

	return rp.get(ctx, sqlf.Sprintf("role_id = %s", opts.RoleID))
}

func (rp *rolePermissionStore) GetByRoleIDAndPermissionID(ctx context.Context, opts GetRolePermissionOpts) (*types.RolePermission, error) {
	if opts.PermissionID == 0 {
		return nil, errors.New("missing permission id")
	}

	if opts.RoleID == 0 {
		return nil, errors.New("missing role id")
	}

	q := sqlf.Sprintf(
		getRolePermissionQueryFmtStr,
		sqlf.Join(rolePermissionColumns, ", "),
		sqlf.Sprintf("role_permissions.role_id = %s AND role_permissions.permission_id = %s", opts.RoleID, opts.PermissionID),
	)

	rolePermission, err := scanRolePermission(rp.QueryRow(ctx, q))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &RolePermissionNotFoundErr{PermissionID: opts.PermissionID, RoleID: opts.RoleID}
		}
		return nil, errors.Wrap(err, "scanning role permission")
	}

	return rolePermission, nil
}

const rolePermissionAssignQueryFmtStr = `
INSERT INTO
	role_permissions (%s)
VALUES %s
ON CONFLICT DO NOTHING
RETURNING %s
`

func (rp *rolePermissionStore) Assign(ctx context.Context, opts AssignRolePermissionOpts) error {
	if opts.PermissionID == 0 {
		return errors.New("missing permission id")
	}

	// First we check that we're not modifying the site administrator role, which
	// should not be modified except by the `UpdatePermissions` startup process.
	err := checkRoleExistsAndIsNotSiteAdminRole(ctx, rp, opts.RoleID)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(
		rolePermissionAssignQueryFmtStr,
		sqlf.Join(rolePermissionInsertColumns, ", "),
		sqlf.Sprintf("( %s, %s )", opts.RoleID, opts.PermissionID),
		sqlf.Join(rolePermissionColumns, ", "),
	)

	_, err = scanRolePermission(rp.QueryRow(ctx, q))
	if err != nil {
		// If there are no rows returned, it means that the role has already being assigned this permission.
		// In that case, we don't need to return an error.
		if err == sql.ErrNoRows {
			return nil
		}
		return errors.Wrap(err, "scanning role permission")
	}
	return nil
}

func (rp *rolePermissionStore) SetPermissionsForRole(ctx context.Context, opts SetPermissionsForRoleOpts) error {
	return rp.WithTransact(ctx, func(tx RolePermissionStore) error {
		// First we check that we're not modifying the site administrator role, which
		// should not be modified except by the `UpdatePermissions` startup process.
		err := checkRoleExistsAndIsNotSiteAdminRole(ctx, tx, opts.RoleID)
		if err != nil {
			return err
		}

		// look up the current permissions assigned to the role. We use this to determine which permissions to assign and revoke.
		rolePermissions, err := tx.GetByRoleID(ctx, GetRolePermissionOpts{RoleID: opts.RoleID})
		if err != nil {
			return err
		}

		// We create a map of permissions for easy lookup.
		var rolePermsMap = make(map[int32]int, len(rolePermissions))
		for _, rolePermission := range rolePermissions {
			rolePermsMap[rolePermission.PermissionID] = 1
		}

		var toBeDeleted []int32
		var toBeAdded []int32

		// figure out permissions that need to be added. Permissions that are received from `opts`
		// and do not exist in the database are new and should be added. While those in the database that aren't
		// part of `opts.Permissions` should be revoked.
		for _, perm := range opts.Permissions {
			count, ok := rolePermsMap[perm]
			if ok {
				// We increment the count of permissions that are in the map, and also part of `opts.Permissions`.
				// These permissions won't be modified.
				rolePermsMap[perm] = count + 1
			} else {
				// Permissions that aren't part of the map (do not exist in the database), should be inserted in the
				// database.
				rolePermsMap[perm] = 0
			}
		}

		// We loop through the map to figure out permissions that should be revoked or assigned.
		for perm, count := range rolePermsMap {
			switch count {
			// Count is zero when the role <> permission association doesn't exist in the database, but is
			// present in `opts.Permissions`.
			case 0:
				toBeAdded = append(toBeAdded, perm)
			// Count is one when the role <> permission association exists in the database, but not in
			// `opts.Permissions`.
			case 1:
				toBeDeleted = append(toBeDeleted, perm)
			}
		}

		// If we have new permissions to be added, we insert into the database via the transaction created earlier.
		if len(toBeAdded) > 0 {
			if err = tx.BulkAssignPermissionsToRole(ctx, BulkAssignPermissionsToRoleOpts{
				RoleID:      opts.RoleID,
				Permissions: toBeAdded,
			}); err != nil {
				return err
			}
		}

		// If we have new permissions to be removed, we remove from the database via the transaction created earlier.
		if len(toBeDeleted) > 0 {
			if err = tx.BulkRevokePermissionsForRole(ctx, BulkRevokePermissionsForRoleOpts{
				RoleID:      opts.RoleID,
				Permissions: toBeDeleted,
			}); err != nil {
				return err
			}
		}

		return nil
	})
}

func (rp *rolePermissionStore) BulkAssignPermissionsToRole(ctx context.Context, opts BulkAssignPermissionsToRoleOpts) error {
	// First we check that we're not modifying the site administrator role, which
	// should not be modified except by the `UpdatePermissions` startup process.
	err := checkRoleExistsAndIsNotSiteAdminRole(ctx, rp, opts.RoleID)
	if err != nil {
		return err
	}

	if len(opts.Permissions) == 0 {
		return errors.New("missing permissions")
	}

	var rps []*sqlf.Query
	for _, permission := range opts.Permissions {
		rps = append(rps, sqlf.Sprintf("( %s, %s )", opts.RoleID, permission))
	}

	q := sqlf.Sprintf(
		rolePermissionAssignQueryFmtStr,
		sqlf.Join(rolePermissionInsertColumns, ", "),
		sqlf.Join(rps, ", "),
		sqlf.Join(rolePermissionColumns, ", "),
	)

	return rp.Exec(ctx, q)
}

func (rp *rolePermissionStore) AssignToSystemRole(ctx context.Context, opts AssignToSystemRoleOpts) error {
	if opts.PermissionID == 0 {
		return errors.New("permission id is required")
	}

	if opts.Role == "" {
		return errors.New("role is required")
	}

	if opts.Role == types.SiteAdministratorSystemRole {
		return errors.New("site administrator role cannot be modified")
	}

	roleQuery := sqlf.Sprintf("SELECT id FROM roles WHERE name = %s", opts.Role)

	q := sqlf.Sprintf(
		rolePermissionAssignQueryFmtStr,
		sqlf.Join(rolePermissionInsertColumns, ", "),
		sqlf.Sprintf("((%s), %s)", roleQuery, opts.PermissionID),
		sqlf.Join(rolePermissionColumns, ", "),
	)

	_, err := scanRolePermission(rp.QueryRow(ctx, q))
	if err != nil {
		// If there are no rows returned, it means that the role has already being assigned this permission.
		// In that case, we don't need to return an error.
		if err == sql.ErrNoRows {
			return nil
		}
		return errors.Wrap(err, "scanning role permission")
	}
	return nil
}

// BulkAssignPermissionsToSystemRoles assigns a permissions to multiple system roles. Note
// that it is the only method allowed to modify the permissions of the
// `SITE_ADMINISTRATOR` role, which it does from the `UpdatePermissions` startup process.
func (rp *rolePermissionStore) BulkAssignPermissionsToSystemRoles(ctx context.Context, opts BulkAssignPermissionsToSystemRolesOpts) error {
	if opts.PermissionID == 0 {
		return errors.New("permission id is required")
	}

	if len(opts.Roles) == 0 {
		return errors.New("roles are required")
	}

	var rps []*sqlf.Query
	for _, role := range opts.Roles {
		roleQuery := sqlf.Sprintf("SELECT id FROM roles WHERE name = %s", role)
		rps = append(rps, sqlf.Sprintf("((%s), %s)", roleQuery, opts.PermissionID))
	}

	q := sqlf.Sprintf(
		rolePermissionAssignQueryFmtStr,
		sqlf.Join(rolePermissionInsertColumns, ", "),
		sqlf.Join(rps, ", "),
		sqlf.Join(rolePermissionColumns, ", "),
	)

	var scanRolePermissions = basestore.NewSliceScanner(scanRolePermission)
	_, err := scanRolePermissions(rp.Query(ctx, q))
	if err != nil {
		// If there are no rows returned, it means that the role has already being assigned this permission.
		// In that case, we don't need to return an error.
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	return nil
}

func checkRoleExistsAndIsNotSiteAdminRole(ctx context.Context, store RolePermissionStore, roleID int32) error {
	if roleID == 0 {
		return errors.New("missing role id")
	}

	roleStore := RolesWith(store)
	role, err := roleStore.Get(ctx, GetRoleOpts{ID: roleID})
	if err != nil {
		return err
	}
	if role == nil {
		return &RoleNotFoundErr{ID: roleID}
	}
	if role.IsSiteAdmin() {
		return errors.New("cannot modify permissions for site admin role")
	}
	return nil
}

type RolePermissionNotFoundErr struct {
	PermissionID int32
	RoleID       int32
}

func (e *RolePermissionNotFoundErr) Error() string {
	return fmt.Sprintf("role permission for role %d and permission %d not found", e.RoleID, e.PermissionID)
}

func (e *RolePermissionNotFoundErr) NotFound() bool {
	return true
}
