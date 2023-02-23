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

var userRoleInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("user_id"),
	sqlf.Sprintf("role_id"),
}

var userRoleColumns = []*sqlf.Query{
	sqlf.Sprintf("user_roles.user_id"),
	sqlf.Sprintf("user_roles.role_id"),
	sqlf.Sprintf("user_roles.created_at"),
}

type UserRoleOpts struct {
	UserID int32
	RoleID int32
}

type (
	AssignUserRoleOpts UserRoleOpts
	RevokeUserRoleOpts UserRoleOpts
	GetUserRoleOpts    UserRoleOpts
)

type AssignSystemRoleOpts struct {
	UserID int32
	Role   types.SystemRole
}

type RevokeSystemRoleOpts struct {
	UserID int32
	Role   types.SystemRole
}

type BulkAssignToUserOpts struct {
	UserID  int32
	RoleIDs []int32
}

type BulkAssignSystemRolesToUserOpts struct {
	UserID int32
	Roles  []types.SystemRole
}

type UserRoleStore interface {
	basestore.ShareableStore

	// Assign is used to assign a role to a user.
	Assign(ctx context.Context, opts AssignUserRoleOpts) error
	// AssignSystemRole assigns a system role to a user.
	AssignSystemRole(ctx context.Context, opts AssignSystemRoleOpts) error
	// BulkAssignToUser assigns multiple roles to a single user. This is useful
	// when we want to assign a user more than one role.
	BulkAssignToUser(ctx context.Context, opts BulkAssignToUserOpts) error
	// BulkAssignToUser assigns multiple system roles to a single user. This is useful
	// when we want to assign a user more than one system role.
	BulkAssignSystemRolesToUser(ctx context.Context, opts BulkAssignSystemRolesToUserOpts) error
	// GetByRoleID returns all UserRole associated with the provided role ID
	GetByRoleID(ctx context.Context, opts GetUserRoleOpts) ([]*types.UserRole, error)
	// GetByRoleIDAndUserID returns one UserRole associated with the provided role and user.
	GetByRoleIDAndUserID(ctx context.Context, opts GetUserRoleOpts) (*types.UserRole, error)
	// GetByUserID returns all UserRole associated with the provided user ID
	GetByUserID(ctx context.Context, opts GetUserRoleOpts) ([]*types.UserRole, error)
	// Revoke deletes the user and role relationship from the database.
	Revoke(ctx context.Context, opts RevokeUserRoleOpts) error
	// RevokeSystemRole revokes a system role that has previously being assigned to a user.
	RevokeSystemRole(ctx context.Context, opts RevokeSystemRoleOpts) error
	// Transact creates a transaction for the UserRoleStore.
	WithTransact(context.Context, func(UserRoleStore) error) error
	// With is used to merge the store with another to pull data via other stores.
	With(basestore.ShareableStore) UserRoleStore
}

type userRoleStore struct {
	*basestore.Store
}

var _ UserRoleStore = &userRoleStore{}

func UserRolesWith(other basestore.ShareableStore) UserRoleStore {
	return &userRoleStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (r *userRoleStore) With(other basestore.ShareableStore) UserRoleStore {
	return &userRoleStore{Store: r.Store.With(other)}
}

func (r *userRoleStore) WithTransact(ctx context.Context, f func(UserRoleStore) error) error {
	return r.Store.WithTransact(ctx, func(tx *basestore.Store) error {
		return f(&userRoleStore{Store: tx})
	})
}

const userRoleAssignQueryFmtStr = `
INSERT INTO
	user_roles (%s)
VALUES %s
ON CONFLICT DO NOTHING
RETURNING %s;
`

func (r *userRoleStore) Assign(ctx context.Context, opts AssignUserRoleOpts) error {
	if opts.UserID == 0 {
		return errors.New("missing user id")
	}

	if opts.RoleID == 0 {
		return errors.New("missing role id")
	}

	q := sqlf.Sprintf(
		userRoleAssignQueryFmtStr,
		sqlf.Join(userRoleInsertColumns, ", "),
		sqlf.Sprintf("( %s, %s )", opts.UserID, opts.RoleID),
		sqlf.Join(userRoleColumns, ", "),
	)

	_, err := scanUserRole(r.QueryRow(ctx, q))
	if err != nil {
		// If there are no rows returned, it means that the user has already being assigned the role.
		// In that case, we don't need to return an error.
		if err == sql.ErrNoRows {
			return nil
		}
		return errors.Wrap(err, "scanning user role")
	}
	return nil
}

func (r *userRoleStore) AssignSystemRole(ctx context.Context, opts AssignSystemRoleOpts) error {
	if opts.UserID == 0 {
		return errors.New("user id is required")
	}

	if opts.Role == "" {
		return errors.New("role is required")
	}

	roleQuery := sqlf.Sprintf("SELECT id FROM roles WHERE name = %s", opts.Role)

	q := sqlf.Sprintf(
		userRoleAssignQueryFmtStr,
		sqlf.Join(userRoleInsertColumns, ", "),
		sqlf.Sprintf("( %s, (%s) )", opts.UserID, roleQuery),
		sqlf.Join(userRoleColumns, ", "),
	)

	_, err := scanUserRole(r.QueryRow(ctx, q))
	if err != nil {
		// If there are no rows returned, it means that the user has already being assigned the role.
		// In that case, we don't need to return an error.
		if err == sql.ErrNoRows {
			return nil
		}
		return errors.Wrap(err, "scanning user role")
	}
	return nil
}

func (r *userRoleStore) BulkAssignToUser(ctx context.Context, opts BulkAssignToUserOpts) error {
	if opts.UserID == 0 {
		return errors.New("missing user id")
	}

	if len(opts.RoleIDs) == 0 {
		return errors.New("missing role ids")
	}

	var urs []*sqlf.Query

	for _, roleId := range opts.RoleIDs {
		urs = append(urs, sqlf.Sprintf("(%s, %s)", opts.UserID, roleId))
	}

	q := sqlf.Sprintf(
		userRoleAssignQueryFmtStr,
		sqlf.Join(userRoleInsertColumns, ", "),
		sqlf.Join(urs, ", "),
		sqlf.Join(userRoleColumns, ", "),
	)

	var scanUserRoles = basestore.NewSliceScanner(scanUserRole)
	_, err := scanUserRoles(r.Query(ctx, q))
	if err != nil {
		// If there are no rows returned, it means that the user has already being assigned the role.
		// In that case, we don't need to return an error.
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	return nil
}

func (r *userRoleStore) BulkAssignSystemRolesToUser(ctx context.Context, opts BulkAssignSystemRolesToUserOpts) error {
	if opts.UserID == 0 {
		return errors.New("user id is required")
	}

	if len(opts.Roles) == 0 {
		return errors.New("roles are required")
	}

	var urs []*sqlf.Query
	for _, role := range opts.Roles {
		roleQuery := sqlf.Sprintf("SELECT id FROM roles WHERE name = %s", role)
		urs = append(urs, sqlf.Sprintf("(%s, (%s))", opts.UserID, roleQuery))
	}

	q := sqlf.Sprintf(
		userRoleAssignQueryFmtStr,
		sqlf.Join(userRoleInsertColumns, ", "),
		sqlf.Join(urs, ", "),
		sqlf.Join(userRoleColumns, ", "),
	)

	var scanUserRoles = basestore.NewSliceScanner(scanUserRole)
	_, err := scanUserRoles(r.Query(ctx, q))
	if err != nil {
		// If there are no rows returned, it means that the user has already being assigned the role.
		// In that case, we don't need to return an error.
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	return nil
}

type UserRoleNotFoundErr struct {
	UserID int32
	RoleID int32
	Role   types.SystemRole
}

func (e *UserRoleNotFoundErr) Error() string {
	return fmt.Sprintf("user role for user %d and role %d not found", e.UserID, e.RoleID)
}

func (e *UserRoleNotFoundErr) NotFound() bool {
	return true
}

const revokeUserRoleQueryFmtStr = `
DELETE FROM user_roles
WHERE %s
`

func (r *userRoleStore) Revoke(ctx context.Context, opts RevokeUserRoleOpts) error {
	if opts.UserID == 0 {
		return errors.New("missing user id")
	}

	if opts.RoleID == 0 {
		return errors.New("missing role id")
	}

	q := sqlf.Sprintf(
		revokeUserRoleQueryFmtStr,
		sqlf.Sprintf("user_id = %s AND role_id = %s", opts.UserID, opts.RoleID),
	)

	result, err := r.ExecResult(ctx, q)
	if err != nil {
		return errors.Wrap(err, "running delete query")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "checking deleted rows")
	}

	if rowsAffected == 0 {
		return errors.Wrap(&UserRoleNotFoundErr{
			UserID: opts.UserID,
			RoleID: opts.RoleID,
		}, "failed to revoke user role")
	}

	return nil
}

func (r *userRoleStore) RevokeSystemRole(ctx context.Context, opts RevokeSystemRoleOpts) error {
	if opts.UserID == 0 {
		return errors.New("userID is required")
	}

	if opts.Role == "" {
		return errors.New("role is required")
	}

	roleQuery := sqlf.Sprintf("SELECT id FROM roles WHERE name = %s", opts.Role)

	q := sqlf.Sprintf(
		revokeUserRoleQueryFmtStr,
		sqlf.Sprintf("user_id = %s AND role_id = (%s)", opts.UserID, roleQuery),
	)

	_, err := r.ExecResult(ctx, q)
	if err != nil {
		return errors.Wrap(err, "running delete query")
	}

	return nil
}

func (r *userRoleStore) GetByUserID(ctx context.Context, opts GetUserRoleOpts) ([]*types.UserRole, error) {
	if opts.UserID == 0 {
		return nil, errors.New("missing user id")
	}
	return r.get(ctx, sqlf.Sprintf("user_id = %s", opts.UserID))
}

func (r *userRoleStore) GetByRoleID(ctx context.Context, opts GetUserRoleOpts) ([]*types.UserRole, error) {
	if opts.RoleID == 0 {
		return nil, errors.New("missing role id")
	}
	return r.get(ctx, sqlf.Sprintf("role_id = %s", opts.RoleID))
}

func (r *userRoleStore) GetByRoleIDAndUserID(ctx context.Context, opts GetUserRoleOpts) (*types.UserRole, error) {
	if opts.UserID == 0 {
		return nil, errors.New("missing user id")
	}

	if opts.RoleID == 0 {
		return nil, errors.New("missing role id")
	}

	q := sqlf.Sprintf(
		getUserRoleQueryFmtStr,
		sqlf.Join(userRoleColumns, ", "),
		sqlf.Sprintf("INNER JOIN users ON user_roles.user_id = users.id"),
		sqlf.Sprintf("users.deleted_at IS NULL AND user_roles.role_id = %s AND user_roles.user_id = %s", opts.RoleID, opts.UserID),
	)

	ur, err := scanUserRole(r.QueryRow(ctx, q))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &UserRoleNotFoundErr{UserID: opts.UserID, RoleID: opts.RoleID}
		}
		return nil, errors.Wrap(err, "scanning user role")
	}

	return ur, nil
}

func scanUserRole(sc dbutil.Scanner) (*types.UserRole, error) {
	var rm types.UserRole
	if err := sc.Scan(
		&rm.UserID,
		&rm.RoleID,
		&rm.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &rm, nil
}

const getUserRoleQueryFmtStr = `
SELECT
	%s
FROM user_roles
-- joins
%s
WHERE %s
`

func (r *userRoleStore) get(ctx context.Context, cond *sqlf.Query) ([]*types.UserRole, error) {
	conds := sqlf.Sprintf("%s AND users.deleted_at IS NULL", cond)
	q := sqlf.Sprintf(
		getUserRoleQueryFmtStr,
		sqlf.Join(userRoleColumns, ", "),
		sqlf.Sprintf("INNER JOIN users ON user_roles.user_id = users.id"),
		conds,
	)

	var scanUserRoles = basestore.NewSliceScanner(scanUserRole)
	return scanUserRoles(r.Query(ctx, q))
}
