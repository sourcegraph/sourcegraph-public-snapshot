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
	CreateUserRoleOpts UserRoleOpts
	DeleteUserRoleOpts UserRoleOpts
	GetUserRoleOpts    UserRoleOpts
)

type BulkCreateForUserOpts struct {
	UserID  int32
	RoleIDs []int32
}

type UserRoleStore interface {
	basestore.ShareableStore

	// Create inserts the given user and role relationship into the database.
	Create(ctx context.Context, opts CreateUserRoleOpts) (*types.UserRole, error)
	// BulkCreateForUser assigns multiple roles to a single user. This is useful
	// when we want to assign a user more than one role.
	BulkCreateForUser(ctx context.Context, opts BulkCreateForUserOpts) ([]*types.UserRole, error)
	// GetByRoleID returns all UserRole associated with the provided role ID
	GetByRoleID(ctx context.Context, opts GetUserRoleOpts) ([]*types.UserRole, error)
	// GetByRoleIDAndUserID returns one UserRole associated with the provided role and user.
	GetByRoleIDAndUserID(ctx context.Context, opts GetUserRoleOpts) (*types.UserRole, error)
	// GetByUserID returns all UserRole associated with the provided user ID
	GetByUserID(ctx context.Context, opts GetUserRoleOpts) ([]*types.UserRole, error)
	// Delete deletes the user and role relationship from the database.
	Delete(ctx context.Context, opts DeleteUserRoleOpts) error
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

const userRoleCreateQueryFmtStr = `
INSERT INTO
	user_roles (%s)
VALUES %s
RETURNING %s;
`

func (r *userRoleStore) Create(ctx context.Context, opts CreateUserRoleOpts) (*types.UserRole, error) {
	if opts.UserID == 0 {
		return nil, errors.New("missing user id")
	}

	if opts.RoleID == 0 {
		return nil, errors.New("missing role id")
	}

	q := sqlf.Sprintf(
		userRoleCreateQueryFmtStr,
		sqlf.Join(userRoleInsertColumns, ", "),
		sqlf.Sprintf("( %s, %s )", opts.UserID, opts.RoleID),
		sqlf.Join(userRoleColumns, ", "),
	)

	rm, err := scanUserRole(r.QueryRow(ctx, q))
	if err != nil {
		return nil, errors.Wrap(err, "scanning user role")
	}
	return rm, nil
}

func (r *userRoleStore) BulkCreateForUser(ctx context.Context, opts BulkCreateForUserOpts) ([]*types.UserRole, error) {
	if opts.UserID == 0 {
		return nil, errors.New("missing user id")
	}

	if len(opts.RoleIDs) == 0 {
		return nil, errors.New("missing role ids")
	}

	var urs []*sqlf.Query

	for _, roleId := range opts.RoleIDs {
		urs = append(urs, sqlf.Sprintf("(%s, %s)", opts.UserID, roleId))
	}

	q := sqlf.Sprintf(
		userRoleCreateQueryFmtStr,
		sqlf.Join(userRoleInsertColumns, ", "),
		sqlf.Join(urs, ", "),
		sqlf.Join(userRoleColumns, ", "),
	)

	rows, err := r.Query(ctx, q)
	if err != nil {
		return nil, errors.Wrap(err, "error running query")
	}
	defer rows.Close()

	var userRoles []*types.UserRole
	for rows.Next() {
		ur, err := scanUserRole(rows)
		if err != nil {
			return userRoles, err
		}
		userRoles = append(userRoles, ur)
	}

	return userRoles, nil
}

type UserRoleNotFoundErr struct {
	UserID int32
	RoleID int32
}

func (e *UserRoleNotFoundErr) Error() string {
	return fmt.Sprintf("user role for user %d and role %d not found", e.UserID, e.RoleID)
}

func (e *UserRoleNotFoundErr) NotFound() bool {
	return true
}

const deleteUserRoleQueryFmtStr = `
DELETE FROM user_roles
WHERE %s
`

func (r *userRoleStore) Delete(ctx context.Context, opts DeleteUserRoleOpts) error {
	if opts.UserID == 0 {
		return errors.New("missing user id")
	}

	if opts.RoleID == 0 {
		return errors.New("missing role id")
	}

	q := sqlf.Sprintf(
		deleteUserRoleQueryFmtStr,
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
		return errors.Wrap(&UserRoleNotFoundErr{opts.UserID, opts.RoleID}, "failed to delete user role")
	}

	return nil
}

func (r *userRoleStore) GetByUserID(ctx context.Context, opts GetUserRoleOpts) ([]*types.UserRole, error) {
	if opts.UserID == 0 {
		return nil, errors.New("missing user id")
	}
	var urs []*types.UserRole

	scanFunc := func(rows *sql.Rows) error {
		ur, err := scanUserRole(rows)
		if err != nil {
			return err
		}
		urs = append(urs, ur)
		return nil
	}

	err := r.get(ctx, sqlf.Sprintf("user_id = %s", opts.UserID), scanFunc)
	return urs, err
}

func (r *userRoleStore) GetByRoleID(ctx context.Context, opts GetUserRoleOpts) ([]*types.UserRole, error) {
	role, err := RolesWith(r).Get(ctx, GetRoleOpts{
		ID: opts.RoleID,
	})
	if err != nil {
		return nil, err
	}

	urs := make([]*types.UserRole, 0, 20)

	scanFunc := func(rows *sql.Rows) error {
		ur, err := scanUserRole(rows)
		if err != nil {
			return err
		}
		urs = append(urs, ur)
		return nil
	}

	err = r.get(ctx, sqlf.Sprintf("role_id = %s", role.ID), scanFunc)
	return urs, err
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

func (r *userRoleStore) get(ctx context.Context, w *sqlf.Query, scanFunc func(rows *sql.Rows) error) error {
	conds := sqlf.Sprintf("%s AND users.deleted_at IS NULL", w)
	q := sqlf.Sprintf(
		getUserRoleQueryFmtStr,
		sqlf.Join(userRoleColumns, ", "),
		sqlf.Sprintf("INNER JOIN users ON user_roles.user_id = users.id"),
		conds,
	)

	rows, err := r.Query(ctx, q)
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
