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

var permissionColumns = []*sqlf.Query{
	sqlf.Sprintf("permissions.id"),
	sqlf.Sprintf("permissions.namespace"),
	sqlf.Sprintf("permissions.action"),
	sqlf.Sprintf("permissions.created_at"),
}

var permissionInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("namespace"),
	sqlf.Sprintf("action"),
}

type PermissionStore interface {
	basestore.ShareableStore

	// Create inserts the given permission into the database.
	Create(ctx context.Context, opts CreatePermissionOpts) (*types.Permission, error)
	// BulkCreate inserts multiple permissions into the database
	BulkCreate(ctx context.Context, opts []CreatePermissionOpts) ([]*types.Permission, error)
	// Delete deletes a permission with the provided ID
	Delete(ctx context.Context, opts DeletePermissionOpts) error
	// GetByID returns the permission matching the given ID, or PermissionNotFoundErr if no such record exists.
	GetByID(ctx context.Context, opts GetPermissionOpts) (*types.Permission, error)
	// List returns all the permissions in the database.
	List(ctx context.Context) ([]*types.Permission, error)
}

type CreatePermissionOpts struct {
	Namespace string
	Action    string
}

type PermissionOpts struct {
	ID int32
}

type (
	GetPermissionOpts    PermissionOpts
	DeletePermissionOpts PermissionOpts
)

type PermissionNotFoundErr struct {
	ID int32
}

func (p *PermissionNotFoundErr) Error() string {
	return fmt.Sprintf("permission with ID %d not found", p.ID)
}

func (p *PermissionNotFoundErr) NotFound() bool {
	return true
}

type permissionStore struct {
	*basestore.Store
}

var _ PermissionStore = &permissionStore{}

const permissionCreateQueryFmtStr = `
INSERT INTO
	permissions(%s)
VALUES %S
RETURNING %s
`

func (p *permissionStore) Create(ctx context.Context, opts CreatePermissionOpts) (*types.Permission, error) {
	q := sqlf.Sprintf(
		permissionCreateQueryFmtStr,
		sqlf.Join(permissionInsertColumns, ", "),
		sqlf.Sprintf("(%s, %s)", opts.Namespace, opts.Action),
		sqlf.Join(permissionColumns, ", "),
	)

	permission, err := scanPermission(p.QueryRow(ctx, q))
	if err != nil {
		return nil, errors.Wrap(err, "scanning role")
	}

	return permission, nil
}

func scanPermission(sc dbutil.Scanner) (*types.Permission, error) {
	var perm types.Permission
	if err := sc.Scan(
		&perm.ID,
		&perm.Namespace,
		&perm.Action,
		&perm.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &perm, nil
}

func (p *permissionStore) BulkCreate(ctx context.Context, opts []CreatePermissionOpts) ([]*types.Permission, error) {
	var values []*sqlf.Query
	for _, opt := range opts {
		values = append(values, sqlf.Sprintf("(%s, %s)", opt.Namespace, opt.Action))
	}

	q := sqlf.Sprintf(
		permissionCreateQueryFmtStr,
		sqlf.Join(permissionInsertColumns, ", "),
		sqlf.Join(values, ", "),
		sqlf.Join(permissionColumns, ", "),
	)

	var perms []*types.Permission
	rows, err := p.Query(ctx, q)
	if err != nil {
		return nil, errors.Wrap(err, "error running query")
	}
	defer rows.Close()
	for rows.Next() {
		perm, err := scanPermission(rows)
		if err != nil {
			return nil, err
		}
		perms = append(perms, perm)
	}

	return perms, rows.Err()
}

const permissionDeleteQueryFmtStr = `
DELETE FROM permissions
WHERE %s
`

func (p *permissionStore) Delete(ctx context.Context, opts DeletePermissionOpts) error {
	if opts.ID == 0 {
		return errors.New("missing id from sql query")
	}

	q := sqlf.Sprintf(permissionDeleteQueryFmtStr, sqlf.Sprintf("id = %s", opts.ID))
	result, err := p.ExecResult(ctx, q)
	if err != nil {
		return errors.Wrap(err, "running delete query")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "checking deleted rows")
	}

	if rowsAffected == 0 {
		return errors.Wrap(&RoleNotFoundErr{opts.ID}, "failed to delete permission")
	}
	return nil
}

const getPermissionQueryFmtStr = `
SELECT %s FROM permissions
WHERE id = %s;
`

func (p *permissionStore) GetByID(ctx context.Context, opts GetPermissionOpts) (*types.Permission, error) {
	if opts.ID == 0 {
		return nil, errors.New("missing id from sql query")
	}

	q := sqlf.Sprintf(
		getPermissionQueryFmtStr,
		sqlf.Join(permissionColumns, ", "),
		opts.ID,
	)

	permission, err := scanPermission(p.QueryRow(ctx, q))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &PermissionNotFoundErr{ID: opts.ID}
		}
		return nil, errors.Wrap(err, "scanning permission")
	}

	return permission, nil
}

const permissionListQueryFmtStr = `
SELECT * FROM permissions
ORDER BY permissions.namespace ASC
`

func (p *permissionStore) List(ctx context.Context) ([]*types.Permission, error) {
	var permissions []*types.Permission
	rows, err := p.Query(ctx, sqlf.Sprintf(permissionListQueryFmtStr))
	if err != nil {
		return nil, errors.Wrap(err, "error running query")
	}

	defer rows.Close()
	for rows.Next() {
		perm, err := scanPermission(rows)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}

	return permissions, rows.Err()
}
