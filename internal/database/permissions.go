package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	rtypes "github.com/sourcegraph/sourcegraph/internal/rbac/types"
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

	// WithTransact creates a transaction-enabled store for the permissionStore
	WithTransact(context.Context, func(PermissionStore) error) error

	// BulkCreate inserts multiple permissions into the database
	BulkCreate(ctx context.Context, opts []CreatePermissionOpts) ([]*types.Permission, error)
	// BulkDelete deletes a permission with the provided ID
	BulkDelete(ctx context.Context, opts []DeletePermissionOpts) error
	// GetPermissionForUser retrieves a permission for a user. If the user doesn't have the permission
	// it returns an error.
	GetPermissionForUser(ctx context.Context, opts GetPermissionForUserOpts) (*types.Permission, error)
	// Count returns the number of permissions in the database matching the options provided.
	Count(ctx context.Context, opts PermissionListOpts) (int, error)
	// Create inserts the given permission into the database.
	Create(ctx context.Context, opts CreatePermissionOpts) (*types.Permission, error)
	// Delete deletes a permission with the provided ID
	Delete(ctx context.Context, opts DeletePermissionOpts) error
	// GetByID returns the permission matching the given ID, or PermissionNotFoundErr if no such record exists.
	GetByID(ctx context.Context, opts GetPermissionOpts) (*types.Permission, error)
	// List returns all the permissions in the database that matches the options.
	List(ctx context.Context, opts PermissionListOpts) ([]*types.Permission, error)
}

type GetPermissionForUserOpts struct {
	UserID int32

	Namespace rtypes.PermissionNamespace
	Action    rtypes.NamespaceAction
}

type CreatePermissionOpts struct {
	Namespace rtypes.PermissionNamespace
	Action    rtypes.NamespaceAction
}

type PermissionOpts struct {
	ID int32
}

type (
	GetPermissionOpts    PermissionOpts
	DeletePermissionOpts PermissionOpts
)

type PermissionListOpts struct {
	PaginationArgs *PaginationArgs

	RoleID int32
	UserID int32
}

type PermissionNotFoundErr struct {
	ID int32

	Namespace rtypes.PermissionNamespace
	Action    rtypes.NamespaceAction
}

func (p *PermissionNotFoundErr) Error() string {
	if p.ID == 0 {
		return fmt.Sprintf("permission %s#%s not found for user", p.Namespace, p.Action)
	}
	return fmt.Sprintf("permission with ID %d not found", p.ID)
}

func (p *PermissionNotFoundErr) NotFound() bool {
	return true
}

type permissionStore struct {
	*basestore.Store
}

var _ PermissionStore = &permissionStore{}

func PermissionsWith(other basestore.ShareableStore) PermissionStore {
	return &permissionStore{Store: basestore.NewWithHandle(other.Handle())}
}

const getPermissionForUserQuery = `
SELECT %s FROM permissions
INNER JOIN role_permissions ON role_permissions.permission_id = permissions.id
INNER JOIN user_roles ON user_roles.role_id = role_permissions.role_id
WHERE permissions.action = %s AND permissions.namespace = %s AND user_roles.user_id = %s
LIMIT 1
`

func (p *permissionStore) GetPermissionForUser(ctx context.Context, opts GetPermissionForUserOpts) (*types.Permission, error) {
	if opts.UserID == 0 {
		return nil, errors.New("missing user id")
	}

	if !opts.Namespace.Valid() {
		return nil, errors.New("invalid permission namespace")
	}

	if opts.Action == "" {
		return nil, errors.New("missing permission action")
	}

	q := sqlf.Sprintf(
		getPermissionForUserQuery,
		sqlf.Join(permissionColumns, ", "),
		opts.Action,
		opts.Namespace,
		opts.UserID,
	)

	permission, err := scanPermission(p.QueryRow(ctx, q))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &PermissionNotFoundErr{
				Namespace: opts.Namespace,
				Action:    opts.Action,
			}
		}
		return nil, errors.Wrap(err, "scanning permission")
	}

	return permission, nil
}

const permissionCreateQueryFmtStr = `
INSERT INTO
	permissions(%s)
VALUES %S
RETURNING %s
`

func (p *permissionStore) WithTransact(ctx context.Context, f func(PermissionStore) error) error {
	return p.Store.WithTransact(ctx, func(tx *basestore.Store) error {
		return f(&permissionStore{Store: tx})
	})
}

func (p *permissionStore) Create(ctx context.Context, opts CreatePermissionOpts) (*types.Permission, error) {
	if opts.Action == "" || !opts.Namespace.Valid() {
		return nil, errors.New("valid action and namespace is required")
	}

	q := sqlf.Sprintf(
		permissionCreateQueryFmtStr,
		sqlf.Join(permissionInsertColumns, ", "),
		sqlf.Sprintf("(%s, %s)", opts.Namespace, opts.Action),
		sqlf.Join(permissionColumns, ", "),
	)

	permission, err := scanPermission(p.QueryRow(ctx, q))
	if err != nil {
		return nil, errors.Wrap(err, "scanning permission")
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
		if !opt.Namespace.Valid() {
			return nil, errors.New("valid namespace is required")
		}
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
		return errors.Wrap(&PermissionNotFoundErr{ID: opts.ID}, "failed to delete permission")
	}
	return nil
}

func (p *permissionStore) BulkDelete(ctx context.Context, opts []DeletePermissionOpts) error {
	if len(opts) == 0 {
		return errors.New("missing ids from sql query")
	}

	var ids []*sqlf.Query
	for _, opt := range opts {
		ids = append(ids, sqlf.Sprintf("%s", opt.ID))
	}

	q := sqlf.Sprintf(
		permissionDeleteQueryFmtStr,
		sqlf.Sprintf(
			"id IN (%s)",
			sqlf.Join(ids, ", "),
		),
	)
	result, err := p.ExecResult(ctx, q)
	if err != nil {
		return errors.Wrap(err, "running delete query")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "checking deleted rows")
	}

	if rowsAffected == 0 {
		return errors.New("failed to delete permissions")
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
SELECT %s FROM permissions
%s
WHERE %s
`

func (p *permissionStore) List(ctx context.Context, opts PermissionListOpts) ([]*types.Permission, error) {
	var permissions []*types.Permission

	scanFunc := func(rows *sql.Rows) error {
		permission, err := scanPermission(rows)
		if err != nil {
			return errors.Wrap(err, "scanning permission")
		}
		permissions = append(permissions, permission)
		return nil
	}

	err := p.list(ctx, opts, scanFunc)
	return permissions, err
}

func (p *permissionStore) list(ctx context.Context, opts PermissionListOpts, scanFunc func(rows *sql.Rows) error) error {
	conds, joins := p.computeConditionsAndJoins(opts)

	queryArgs := opts.PaginationArgs.SQL()
	if queryArgs.Where != nil {
		conds = append(conds, queryArgs.Where)
	}

	if len(conds) == 0 {
		conds = append(conds, sqlf.Sprintf("TRUE"))
	}

	query := sqlf.Sprintf(
		permissionListQueryFmtStr,
		sqlf.Join(permissionColumns, ", "),
		joins,
		sqlf.Join(conds, "AND "),
	)

	if opts.UserID != 0 {
		// We group by `permissions.id` because it's possible for a user to have multiple occurrences of a particular
		// permission. We only want the distinct permissions assigned to a user.
		query = sqlf.Sprintf("%s\n%s", query, sqlf.Sprintf("GROUP BY permissions.id"))
	}

	query = queryArgs.AppendOrderToQuery(query)
	query = queryArgs.AppendLimitToQuery(query)

	rows, err := p.Query(ctx, query)
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

func (p *permissionStore) computeConditionsAndJoins(opts PermissionListOpts) ([]*sqlf.Query, *sqlf.Query) {
	conds := []*sqlf.Query{}
	joins := sqlf.Sprintf("")

	if opts.RoleID != 0 {
		conds = append(conds, sqlf.Sprintf("role_permissions.role_id = %s", opts.RoleID))
		joins = sqlf.Sprintf("INNER JOIN role_permissions ON role_permissions.permission_id = permissions.id")
	}

	if opts.UserID != 0 {
		conds = append(conds, sqlf.Sprintf("user_roles.user_id = %s", opts.UserID))
		joins = sqlf.Sprintf(`
INNER JOIN role_permissions ON role_permissions.permission_id = permissions.id
INNER JOIN user_roles ON user_roles.role_id = role_permissions.role_id
`)
	}

	return conds, joins
}

const permissionCountQueryFmtstr = `
SELECT COUNT(DISTINCT id) FROM permissions
%s
WHERE %s
`

func (p *permissionStore) Count(ctx context.Context, opts PermissionListOpts) (c int, err error) {
	conds, joins := p.computeConditionsAndJoins(opts)

	if len(conds) == 0 {
		conds = append(conds, sqlf.Sprintf("TRUE"))
	}

	query := sqlf.Sprintf(
		permissionCountQueryFmtstr,
		joins,
		sqlf.Join(conds, " AND "),
	)

	count, _, err := basestore.ScanFirstInt(p.Query(ctx, query))
	return count, err
}
