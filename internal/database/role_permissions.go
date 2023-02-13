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
	CreateRolePermissionOpts RolePermissionOpts
	DeleteRolePermissionOpts RolePermissionOpts
	GetRolePermissionOpts    RolePermissionOpts
)

type RolePermissionStore interface {
	basestore.ShareableStore

	// Create inserts the given role and permission relationship into the database.
	Create(ctx context.Context, opts CreateRolePermissionOpts) (*types.RolePermission, error)
	// GetByRoleIDAndPermissionID returns one RolePermission associated with the provided role and permission.
	GetByRoleIDAndPermissionID(ctx context.Context, opts GetRolePermissionOpts) (*types.RolePermission, error)
	// GetByRoleID returns all RolePermission associated with the provided role ID
	GetByRoleID(ctx context.Context, opts GetRolePermissionOpts) ([]*types.RolePermission, error)
	// GetByPermissionID returns all RolePermission associated with the provided permission ID
	GetByPermissionID(ctx context.Context, opts GetRolePermissionOpts) ([]*types.RolePermission, error)
	// Delete deletes the permission and role relationship from the database.
	Delete(ctx context.Context, opts DeleteRolePermissionOpts) error
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

func (rp *rolePermissionStore) Delete(ctx context.Context, opts DeleteRolePermissionOpts) error {
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
		return errors.Wrap(&RolePermissionNotFoundErr{opts.PermissionID, opts.RoleID}, "failed to delete role permission")
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

func (rp *rolePermissionStore) get(ctx context.Context, w *sqlf.Query, scanFunc func(rows *sql.Rows) error) error {
	q := sqlf.Sprintf(
		getRolePermissionQueryFmtStr,
		sqlf.Join(rolePermissionColumns, ", "),
		w,
	)

	fmt.Println(q.Query(sqlf.PostgresBindVar))

	rows, err := rp.Query(ctx, q)
	if err != nil {
		return errors.Wrap(err, "error running query")
	}
	defer rows.Close()
	for rows.Next() {
		if err := scanFunc(rows); err != nil {
			return err
		}
	}

	return rows.Err()
}

func (rp *rolePermissionStore) GetByPermissionID(ctx context.Context, opts GetRolePermissionOpts) ([]*types.RolePermission, error) {
	if opts.PermissionID == 0 {
		return nil, errors.New("missing permission id")
	}

	var rps []*types.RolePermission

	scanFunc := func(rows *sql.Rows) error {
		rp, err := scanRolePermission(rows)
		if err != nil {
			return err
		}
		rps = append(rps, rp)
		return nil
	}

	err := rp.get(ctx, sqlf.Sprintf("permission_id = %s", opts.PermissionID), scanFunc)
	return rps, err
}

func (rp *rolePermissionStore) GetByRoleID(ctx context.Context, opts GetRolePermissionOpts) ([]*types.RolePermission, error) {
	if opts.RoleID == 0 {
		return nil, errors.New("missing role id")
	}

	var rps []*types.RolePermission

	scanFunc := func(rows *sql.Rows) error {
		rp, err := scanRolePermission(rows)
		if err != nil {
			return err
		}
		rps = append(rps, rp)
		return nil
	}

	err := rp.get(ctx, sqlf.Sprintf("role_id = %s", opts.RoleID), scanFunc)
	return rps, err
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

const rolePermissionCreateQueryFmtStr = `
INSERT INTO
	role_permissions (%s)
VALUES (
	%s,
	%s
)
RETURNING %s
`

func (rp *rolePermissionStore) Create(ctx context.Context, opts CreateRolePermissionOpts) (*types.RolePermission, error) {
	if opts.PermissionID == 0 {
		return nil, errors.New("missing permission id")
	}

	if opts.RoleID == 0 {
		return nil, errors.New("missing role id")
	}

	q := sqlf.Sprintf(
		rolePermissionCreateQueryFmtStr,
		sqlf.Join(rolePermissionInsertColumns, ", "),
		opts.RoleID,
		opts.PermissionID,
		sqlf.Join(rolePermissionColumns, ", "),
	)

	rolePermission, err := scanRolePermission(rp.QueryRow(ctx, q))
	if err != nil {
		return nil, errors.Wrap(err, "scanning user role")
	}
	return rolePermission, nil
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
