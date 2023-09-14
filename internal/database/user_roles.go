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

type BulkAssignSystemRolesToUserOpts struct {
	UserID int32
	Roles  []types.SystemRole
}

type BulkUserRoleOperationOpts struct {
	UserID int32
	Roles  []int32
}

type (
	BulkAssignRolesToUserOpts  BulkUserRoleOperationOpts
	SetRolesForUserOpts        BulkUserRoleOperationOpts
	BulkRevokeRolesForUserOpts BulkUserRoleOperationOpts
)

type UserRoleStore interface {
	basestore.ShareableStore

	// Assign is used to assign a role to a user.
	Assign(ctx context.Context, opts AssignUserRoleOpts) error
	// AssignSystemRole assigns a system role to a user.
	AssignSystemRole(ctx context.Context, opts AssignSystemRoleOpts) error
	// BulkAssignRolesToUser assigns multiple roles to a single user. This is useful
	// when we want to assign a user more than one role.
	BulkAssignRolesToUser(ctx context.Context, opts BulkAssignRolesToUserOpts) error
	// BulkAssignRolesToUser assigns multiple system roles to a single user. This is useful
	// when we want to assign a user more than one system role.
	BulkAssignSystemRolesToUser(ctx context.Context, opts BulkAssignSystemRolesToUserOpts) error
	// BulkRevokeRolesForUser revokes bulk roles assigned to a user.
	BulkRevokeRolesForUser(ctx context.Context, opts BulkRevokeRolesForUserOpts) error
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
	// SetRolesForUser is used to sync the roles assigned to a role. It removes any role that isn't
	// included in the `opts` and assigns roles that aren't yet assigned in the database but in `opts`.
	SetRolesForUser(ctx context.Context, opts SetRolesForUserOpts) error
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

func (r *userRoleStore) BulkAssignRolesToUser(ctx context.Context, opts BulkAssignRolesToUserOpts) error {
	if opts.UserID == 0 {
		return errors.New("missing user id")
	}

	if len(opts.Roles) == 0 {
		return errors.New("missing role ids")
	}

	var urs []*sqlf.Query

	for _, roleId := range opts.Roles {
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

func (r *userRoleStore) BulkRevokeRolesForUser(ctx context.Context, opts BulkRevokeRolesForUserOpts) error {
	if opts.UserID == 0 {
		return errors.New("missing user id")
	}

	if len(opts.Roles) == 0 {
		return errors.New("missing roles")
	}

	var preds []*sqlf.Query
	var roleIDs []*sqlf.Query

	for _, role := range opts.Roles {
		roleIDs = append(roleIDs, sqlf.Sprintf("%s", role))
	}

	preds = append(preds, sqlf.Sprintf("user_id = %s", opts.UserID))
	preds = append(preds, sqlf.Sprintf("role_id IN ( %s )", sqlf.Join(roleIDs, ", ")))

	q := sqlf.Sprintf(
		revokeUserRoleQueryFmtStr,
		sqlf.Join(preds, " AND "),
	)

	return r.Exec(ctx, q)
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

func (r *userRoleStore) SetRolesForUser(ctx context.Context, opts SetRolesForUserOpts) error {
	if opts.UserID == 0 {
		return errors.New("missing user id")
	}

	return r.WithTransact(ctx, func(tx UserRoleStore) error {
		// look up the current roles assigned to the user. We use this to determine which roles to assign and revoke.
		userRoles, err := tx.GetByUserID(ctx, GetUserRoleOpts{UserID: opts.UserID})
		if err != nil {
			return err
		}

		// We create a map of roles for easy lookup.
		var userRolesMap = make(map[int32]int, len(userRoles))
		for _, userRole := range userRoles {
			userRolesMap[userRole.RoleID] = 1
		}

		var toBeDeleted []int32
		var toBeAdded []int32

		// figure out rikes that need to be added. Roles that are received from `opts`
		// and do not exist in the database are new and should be added. While those in the database that aren't
		// part of `opts.Roles` should be revoked.
		for _, role := range opts.Roles {
			count, ok := userRolesMap[role]
			if ok {
				// We increment the count of roles that are in the map, and also part of `opts.Roles`.
				// These roles won't be modified.
				userRolesMap[role] = count + 1
			} else {
				// Roles that aren't part of the map (do not exist in the database), should be inserted in the
				// database.
				userRolesMap[role] = 0
			}
		}

		// We loop through the map to figure out roles that should be revoked or assigned.
		for role, count := range userRolesMap {
			switch count {
			// Count is zero when the user <> role association doesn't exist in the database, but is
			// present in `opts.Roles`.
			case 0:
				toBeAdded = append(toBeAdded, role)
			// Count is one when the user <> role association exists in the database, but not in
			// `opts.Roles`.
			case 1:
				toBeDeleted = append(toBeDeleted, role)
			}
		}

		// If we have new permissions to be added, we insert into the database via the transaction created earlier.
		if len(toBeAdded) > 0 {
			if err = tx.BulkAssignRolesToUser(ctx, BulkAssignRolesToUserOpts{
				UserID: opts.UserID,
				Roles:  toBeAdded,
			}); err != nil {
				return err
			}
		}

		// If we have new permissions to be removed, we remove from the database via the transaction created earlier.
		if len(toBeDeleted) > 0 {
			if err = tx.BulkRevokeRolesForUser(ctx, BulkRevokeRolesForUserOpts{
				UserID: opts.UserID,
				Roles:  toBeDeleted,
			}); err != nil {
				return err
			}
		}

		return nil
	})
}
