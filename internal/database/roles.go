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

var roleColumns = []*sqlf.Query{
	sqlf.Sprintf("roles.id"),
	sqlf.Sprintf("roles.name"),
	sqlf.Sprintf("roles.readonly"),
	sqlf.Sprintf("roles.created_at"),
	sqlf.Sprintf("roles.deleted_at"),
}

var roleInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("name"),
	sqlf.Sprintf("readonly"),
}

var (
	DefaultRole           string = "DEFAULT"
	SiteAdministratorRole string = "SITE_ADMINISTRATOR"
)

type RoleStore interface {
	basestore.ShareableStore

	// Count counts all roles in the database.
	Count(ctx context.Context, opts RolesListOptions) (int, error)
	// Create inserts the given role into the database.
	Create(ctx context.Context, name string, readonly bool) (*types.Role, error)
	// Delete removes an existing role from the database.
	Delete(ctx context.Context, opts DeleteRoleOpts) error
	// Get returns the role matching the given ID or name provided. If no such role exists,
	// a RoleNotFoundErr is returned.
	Get(ctx context.Context, opts GetRoleOpts) (*types.Role, error)
	// GetByID returns the role matching the given ID, or RoleNotFoundErr if no such record exists.
	GetByID(ctx context.Context, opts GetRoleOpts) (*types.Role, error)
	// List returns all roles matching the given options.
	List(ctx context.Context, opts RolesListOptions) ([]*types.Role, error)
	// Update updates an existing role in the database.
	Update(ctx context.Context, role *types.Role) (*types.Role, error)
}

func RolesWith(other basestore.ShareableStore) RoleStore {
	return &roleStore{Store: basestore.NewWithHandle(other.Handle())}
}

type RoleOpts struct {
	ID   int32
	Name string
}

type (
	DeleteRoleOpts RoleOpts
	GetRoleOpts    RoleOpts
)

type RolesListOptions struct {
	*LimitOffset
	ReadOnly bool
}

type RoleNotFoundErr struct {
	ID int32
}

func (e *RoleNotFoundErr) Error() string {
	return fmt.Sprintf("role with ID %d not found", e.ID)
}

func (e *RoleNotFoundErr) NotFound() bool {
	return true
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

func (r *roleStore) Get(ctx context.Context, opts GetRoleOpts) (*types.Role, error) {
	if opts.ID == 0 && opts.Name == "" {
		return nil, errors.New("missing id or name")
	}

	if opts.ID > 0 {
		return r.GetByID(ctx, opts)
	}

	whereClause := sqlf.Sprintf("name = %s AND deleted_at IS NULL", opts.Name)
	q := sqlf.Sprintf(
		getRoleFmtStr,
		sqlf.Join(roleColumns, ", "),
		whereClause,
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

func (r *roleStore) GetByID(ctx context.Context, opts GetRoleOpts) (*types.Role, error) {
	if opts.ID <= 0 {
		return nil, errors.New("missing id from sql query")
	}

	whereClause := sqlf.Sprintf("id = %s AND deleted_at IS NULL", opts.ID)

	q := sqlf.Sprintf(
		getRoleFmtStr,
		sqlf.Join(roleColumns, ", "),
		whereClause,
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

func (r *roleStore) List(ctx context.Context, opts RolesListOptions) ([]*types.Role, error) {
	roles := make([]*types.Role, 0, 20)

	scanFunc := func(rows *sql.Rows) error {
		role, err := scanRole(rows)
		if err != nil {
			return err
		}
		roles = append(roles, role)
		return nil
	}

	err := r.list(ctx, opts, sqlf.Join(roleColumns, ", "), sqlf.Sprintf("ORDER BY roles.created_at ASC"), scanFunc)
	return roles, err
}

const roleListQueryFmtstr = `
SELECT
	%s
FROM roles
WHERE %s
%s
`

func (r *roleStore) list(ctx context.Context, opts RolesListOptions, selects, orderByQuery *sqlf.Query, scanRole func(rows *sql.Rows) error) error {
	var whereClause = []*sqlf.Query{sqlf.Sprintf("deleted_at IS NULL")}

	if opts.ReadOnly {
		whereClause = append(whereClause, sqlf.Sprintf("readonly IS TRUE"))
	}

	q := sqlf.Sprintf(roleListQueryFmtstr, selects, sqlf.Join(whereClause, " AND "), orderByQuery)

	if opts.LimitOffset != nil {
		q = sqlf.Sprintf("%s\n%s", q, opts.LimitOffset.SQL())
	}

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
	roles (%s)
	VALUES (
		%s,
		%s
	)
	RETURNING %s

`

func (r *roleStore) Create(ctx context.Context, name string, readonly bool) (_ *types.Role, err error) {
	q := sqlf.Sprintf(
		roleCreateQueryFmtStr,
		sqlf.Join(roleInsertColumns, ", "),
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

func (r *roleStore) Count(ctx context.Context, opts RolesListOptions) (c int, err error) {
	opts.LimitOffset = nil
	err = r.list(ctx, opts, sqlf.Sprintf("COUNT(1)"), sqlf.Sprintf(""), func(rows *sql.Rows) error {
		return rows.Scan(&c)
	})
	return c, err
}

const roleUpdateQueryFmtstr = `
UPDATE roles
SET
    name = %s
WHERE
	id = %s
RETURNING
	%s
`

func (r *roleStore) Update(ctx context.Context, role *types.Role) (*types.Role, error) {
	q := sqlf.Sprintf(roleUpdateQueryFmtstr, role.Name, role.ID, sqlf.Join(roleColumns, ", "))

	updated, err := scanRole(r.QueryRow(ctx, q))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &RoleNotFoundErr{ID: role.ID}
		}
		return nil, errors.Wrap(err, "scanning role")
	}
	return updated, nil
}

const roleDeleteQueryFmtStr = `
UPDATE roles
SET
	deleted_at = NOW()
WHERE id = %s AND NOT readonly
`

func (r *roleStore) Delete(ctx context.Context, opts DeleteRoleOpts) error {
	if opts.ID <= 0 {
		return errors.New("missing id from sql query")
	}

	// We don't allow deletion of readonly roles such as DEFAULT & SITE_ADMINISTRATOR
	q := sqlf.Sprintf(roleDeleteQueryFmtStr, opts.ID)
	result, err := r.ExecResult(ctx, q)
	if err != nil {
		return errors.Wrap(err, "running delete query")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "checking deleted rows")
	}

	if rowsAffected == 0 {
		return errors.Wrap(&RoleNotFoundErr{opts.ID}, "failed to delete role")
	}
	return nil
}
