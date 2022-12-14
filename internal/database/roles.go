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

const defaultRoleLimit = 20

var roleColumns = []*sqlf.Query{
	sqlf.Sprintf("roles.id"),
	sqlf.Sprintf("roles.name"),
	sqlf.Sprintf("roles.readonly"),
	sqlf.Sprintf("roles.created_at"),
	sqlf.Sprintf("roles.deleted_at"),
}

var permissionColumnsForRole = []*sqlf.Query{
	sqlf.Sprintf("permissions.id AS permission_id"),
	sqlf.Sprintf("permissions.action"),
	sqlf.Sprintf("permissions.namespace"),
}

var roleInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("name"),
	sqlf.Sprintf("readonly"),
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("deleted_at"),
}

type RoleStore interface {
	basestore.ShareableStore

	GetByID(context.Context, GetRoleOpts) (*types.Role, error)
	List(context.Context, RolesListOptions) ([]*types.Role, error)
	Create(context.Context, string, bool) (*types.Role, error)
}

type RoleOpts struct {
	ID int32
}

type (
	DeleteRoleOpts RoleOpts
	GetRoleOpts    RoleOpts
)

type RolesListOptions struct {
	*LimitOffset
	Cursor         *types.Cursor
	IncludeDeleted bool
}

type RoleNotFoundErr struct {
	ID int32
}

func (e *RoleNotFoundErr) Error() string {
	return fmt.Sprintf("role with ID %d not found", e.ID)
}

type roleStore struct {
	*basestore.Store
}

var _ RoleStore = &roleStore{}

const getRoleFmtStr = `
SELECT %s FROM roles
WHERE %s
LIMIT 1;
`

func (r *roleStore) GetByID(ctx context.Context, opts GetRoleOpts) (*types.Role, error) {
	var preds []*sqlf.Query

	if opts.ID > 0 {
		preds = append(preds, sqlf.Sprintf("id = %d", opts.ID))
	}

	preds = append(preds, sqlf.Sprintf("deleted_at IS NULL"))

	q := sqlf.Sprintf(
		getRoleFmtStr,
		sqlf.Join(roleColumns, ", "),
		sqlf.Join(preds, "\n AND "),
	)

	role, err := scanRole(r.QueryRow(ctx, q))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &RoleNotFoundErr{ID: opts.ID}
		}
		return nil, errors.Wrap(err, "scanning role")
	}

	return role, nil
}

func scanRole(sc dbutil.Scanner) (*types.Role, error) {
	var role types.Role
	if err := sc.Scan(
		&role.ID,
		&role.Name,
		&role.ReadOnly,
		&role.CreatedAt,
		&dbutil.NullTime{Time: &role.DeletedAt},
	); err != nil {
		return nil, err
	}

	return &role, nil
}

const roleListQueryFmtstr = `
SELECT
	%s
FROM roles
WHERE %s
%s
ORDER BY id ASC;
`

func (r *roleStore) List(ctx context.Context, opts RolesListOptions) ([]*types.Role, error) {
	if opts.Limit == 0 {
		opts.Limit = defaultRoleLimit
	}
	roles := make([]*types.Role, 0, opts.Limit)

	scanFunc := func(rows *sql.Rows) error {
		role, err := scanRole(rows)
		if err != nil {
			return err
		}
		roles = append(roles, role)
		return nil
	}

	err := r.list(ctx, opts, sqlf.Join(roleColumns, ", "), scanFunc)
	return roles, err
}

func (r *roleStore) list(ctx context.Context, opts RolesListOptions, selects *sqlf.Query, scanRole func(rows *sql.Rows) error) error {
	var whereClause *sqlf.Query
	if opts.IncludeDeleted {
		whereClause = sqlf.Sprintf("deleted_at IS NOT NULL")
	} else {
		whereClause = sqlf.Sprintf("deleted_at IS NULL")
	}

	q := sqlf.Sprintf(roleListQueryFmtstr, selects, whereClause, opts.SQL())

	rows, err := r.Query(ctx, q)
	if err != nil {
		return errors.Wrap(err, "error running query")
	}
	defer rows.Close()
	for rows.Next() {
		if err := scanRole(rows); err != nil {
			return err
		}
	}

	return rows.Err()
}

const roleCreateQueryFmtStr = `
INSERT INTO
	roles (
		name,
		readonly
	)
	VALUES (
		%s,
		%s
	)
	RETURNING %s

`

func (r *roleStore) Create(ctx context.Context, name string, readonly bool) (_ *types.Role, err error) {
	q := sqlf.Sprintf(
		roleCreateQueryFmtStr,
		name,
		readonly,
		// Returning
		sqlf.Join(roleColumns, ", "),
	)

	role, err := scanRole(r.QueryRow(ctx, q))
	if err != nil {
		return nil, errors.Wrap(err, "scanning role")
	}

	return role, nil
}
