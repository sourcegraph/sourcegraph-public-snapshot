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

type RolePermissionStore interface {
	basestore.ShareableStore

	// Assign is used to assign a permission to a role.
	Assign(ctx context.Context, opts AssignRolePermissionOpts) error
	// AssignToSystemRole is used to assign a permission to a system role.
	AssignToSystemRole(ctx context.Context, opts AssignToSystemRoleOpts) error
	// BulkAssignToSystemRole is used to assign a permission to multiple system roles.
	BulkAssignPermissionsToSystemRoles(ctx context.Context, opts BulkAssignPermissionsToSystemRolesOpts) error
	// GetByRoleIDAndPermissionID returns one RolePermission associated with the provided role and permission.
	GetByRoleIDAndPermissionID(ctx context.Context, opts GetRolePermissionOpts) (*types.RolePermission, error)
	// GetByRoleID returns all RolePermission associated with the provided role ID
	GetByRoleID(ctx context.Context, opts GetRolePermissionOpts) ([]*types.RolePermission, error)
	// GetByPermissionID returns all RolePermission associated with the provided permission ID
	GetByPermissionID(ctx context.Context, opts GetRolePermissionOpts) ([]*types.RolePermission, error)
	// Revoke deletes the permission and role relationship from the database.
	Revoke(ctx context.Context, opts RevokeRolePermissionOpts) error
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

	if opts.RoleID == 0 {
		return errors.New("missing role id")
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

	if opts.RoleID == 0 {
		return errors.New("missing role id")
	}

	q := sqlf.Sprintf(
		rolePermissionAssignQueryFmtStr,
		sqlf.Join(rolePermissionInsertColumns, ", "),
		sqlf.Sprintf("( %s, %s )", opts.RoleID, opts.PermissionID),
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

func (rp *rolePermissionStore) AssignToSystemRole(ctx context.Context, opts AssignToSystemRoleOpts) error {
	if opts.PermissionID == 0 {
		return errors.New("permission id is required")
	}

	if opts.Role == "" {
		return errors.New("role is required")
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
