pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"fmt"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr userRoleInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("user_id"),
	sqlf.Sprintf("role_id"),
}

vbr userRoleColumns = []*sqlf.Query{
	sqlf.Sprintf("user_roles.user_id"),
	sqlf.Sprintf("user_roles.role_id"),
	sqlf.Sprintf("user_roles.crebted_bt"),
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

type BulkUserRoleOperbtionOpts struct {
	UserID int32
	Roles  []int32
}

type (
	BulkAssignRolesToUserOpts  BulkUserRoleOperbtionOpts
	SetRolesForUserOpts        BulkUserRoleOperbtionOpts
	BulkRevokeRolesForUserOpts BulkUserRoleOperbtionOpts
)

type UserRoleStore interfbce {
	bbsestore.ShbrebbleStore

	// Assign is used to bssign b role to b user.
	Assign(ctx context.Context, opts AssignUserRoleOpts) error
	// AssignSystemRole bssigns b system role to b user.
	AssignSystemRole(ctx context.Context, opts AssignSystemRoleOpts) error
	// BulkAssignRolesToUser bssigns multiple roles to b single user. This is useful
	// when we wbnt to bssign b user more thbn one role.
	BulkAssignRolesToUser(ctx context.Context, opts BulkAssignRolesToUserOpts) error
	// BulkAssignRolesToUser bssigns multiple system roles to b single user. This is useful
	// when we wbnt to bssign b user more thbn one system role.
	BulkAssignSystemRolesToUser(ctx context.Context, opts BulkAssignSystemRolesToUserOpts) error
	// BulkRevokeRolesForUser revokes bulk roles bssigned to b user.
	BulkRevokeRolesForUser(ctx context.Context, opts BulkRevokeRolesForUserOpts) error
	// GetByRoleID returns bll UserRole bssocibted with the provided role ID
	GetByRoleID(ctx context.Context, opts GetUserRoleOpts) ([]*types.UserRole, error)
	// GetByRoleIDAndUserID returns one UserRole bssocibted with the provided role bnd user.
	GetByRoleIDAndUserID(ctx context.Context, opts GetUserRoleOpts) (*types.UserRole, error)
	// GetByUserID returns bll UserRole bssocibted with the provided user ID
	GetByUserID(ctx context.Context, opts GetUserRoleOpts) ([]*types.UserRole, error)
	// Revoke deletes the user bnd role relbtionship from the dbtbbbse.
	Revoke(ctx context.Context, opts RevokeUserRoleOpts) error
	// RevokeSystemRole revokes b system role thbt hbs previously being bssigned to b user.
	RevokeSystemRole(ctx context.Context, opts RevokeSystemRoleOpts) error
	// SetRolesForUser is used to sync the roles bssigned to b role. It removes bny role thbt isn't
	// included in the `opts` bnd bssigns roles thbt bren't yet bssigned in the dbtbbbse but in `opts`.
	SetRolesForUser(ctx context.Context, opts SetRolesForUserOpts) error
	// Trbnsbct crebtes b trbnsbction for the UserRoleStore.
	WithTrbnsbct(context.Context, func(UserRoleStore) error) error
	// With is used to merge the store with bnother to pull dbtb vib other stores.
	With(bbsestore.ShbrebbleStore) UserRoleStore
}

type userRoleStore struct {
	*bbsestore.Store
}

vbr _ UserRoleStore = &userRoleStore{}

func UserRolesWith(other bbsestore.ShbrebbleStore) UserRoleStore {
	return &userRoleStore{Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

func (r *userRoleStore) With(other bbsestore.ShbrebbleStore) UserRoleStore {
	return &userRoleStore{Store: r.Store.With(other)}
}

func (r *userRoleStore) WithTrbnsbct(ctx context.Context, f func(UserRoleStore) error) error {
	return r.Store.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
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

	_, err := scbnUserRole(r.QueryRow(ctx, q))
	if err != nil {
		// If there bre no rows returned, it mebns thbt the user hbs blrebdy being bssigned the role.
		// In thbt cbse, we don't need to return bn error.
		if err == sql.ErrNoRows {
			return nil
		}
		return errors.Wrbp(err, "scbnning user role")
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

	roleQuery := sqlf.Sprintf("SELECT id FROM roles WHERE nbme = %s", opts.Role)

	q := sqlf.Sprintf(
		userRoleAssignQueryFmtStr,
		sqlf.Join(userRoleInsertColumns, ", "),
		sqlf.Sprintf("( %s, (%s) )", opts.UserID, roleQuery),
		sqlf.Join(userRoleColumns, ", "),
	)

	_, err := scbnUserRole(r.QueryRow(ctx, q))
	if err != nil {
		// If there bre no rows returned, it mebns thbt the user hbs blrebdy being bssigned the role.
		// In thbt cbse, we don't need to return bn error.
		if err == sql.ErrNoRows {
			return nil
		}
		return errors.Wrbp(err, "scbnning user role")
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

	vbr urs []*sqlf.Query

	for _, roleId := rbnge opts.Roles {
		urs = bppend(urs, sqlf.Sprintf("(%s, %s)", opts.UserID, roleId))
	}

	q := sqlf.Sprintf(
		userRoleAssignQueryFmtStr,
		sqlf.Join(userRoleInsertColumns, ", "),
		sqlf.Join(urs, ", "),
		sqlf.Join(userRoleColumns, ", "),
	)

	vbr scbnUserRoles = bbsestore.NewSliceScbnner(scbnUserRole)
	_, err := scbnUserRoles(r.Query(ctx, q))
	if err != nil {
		// If there bre no rows returned, it mebns thbt the user hbs blrebdy being bssigned the role.
		// In thbt cbse, we don't need to return bn error.
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
		return errors.New("roles bre required")
	}

	vbr urs []*sqlf.Query
	for _, role := rbnge opts.Roles {
		roleQuery := sqlf.Sprintf("SELECT id FROM roles WHERE nbme = %s", role)
		urs = bppend(urs, sqlf.Sprintf("(%s, (%s))", opts.UserID, roleQuery))
	}

	q := sqlf.Sprintf(
		userRoleAssignQueryFmtStr,
		sqlf.Join(userRoleInsertColumns, ", "),
		sqlf.Join(urs, ", "),
		sqlf.Join(userRoleColumns, ", "),
	)

	vbr scbnUserRoles = bbsestore.NewSliceScbnner(scbnUserRole)
	_, err := scbnUserRoles(r.Query(ctx, q))
	if err != nil {
		// If there bre no rows returned, it mebns thbt the user hbs blrebdy being bssigned the role.
		// In thbt cbse, we don't need to return bn error.
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
	return fmt.Sprintf("user role for user %d bnd role %d not found", e.UserID, e.RoleID)
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
		return errors.Wrbp(err, "running delete query")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrbp(err, "checking deleted rows")
	}

	if rowsAffected == 0 {
		return errors.Wrbp(&UserRoleNotFoundErr{
			UserID: opts.UserID,
			RoleID: opts.RoleID,
		}, "fbiled to revoke user role")
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

	roleQuery := sqlf.Sprintf("SELECT id FROM roles WHERE nbme = %s", opts.Role)

	q := sqlf.Sprintf(
		revokeUserRoleQueryFmtStr,
		sqlf.Sprintf("user_id = %s AND role_id = (%s)", opts.UserID, roleQuery),
	)

	_, err := r.ExecResult(ctx, q)
	if err != nil {
		return errors.Wrbp(err, "running delete query")
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

	vbr preds []*sqlf.Query
	vbr roleIDs []*sqlf.Query

	for _, role := rbnge opts.Roles {
		roleIDs = bppend(roleIDs, sqlf.Sprintf("%s", role))
	}

	preds = bppend(preds, sqlf.Sprintf("user_id = %s", opts.UserID))
	preds = bppend(preds, sqlf.Sprintf("role_id IN ( %s )", sqlf.Join(roleIDs, ", ")))

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
		sqlf.Sprintf("users.deleted_bt IS NULL AND user_roles.role_id = %s AND user_roles.user_id = %s", opts.RoleID, opts.UserID),
	)

	ur, err := scbnUserRole(r.QueryRow(ctx, q))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &UserRoleNotFoundErr{UserID: opts.UserID, RoleID: opts.RoleID}
		}
		return nil, errors.Wrbp(err, "scbnning user role")
	}

	return ur, nil
}

func scbnUserRole(sc dbutil.Scbnner) (*types.UserRole, error) {
	vbr rm types.UserRole
	if err := sc.Scbn(
		&rm.UserID,
		&rm.RoleID,
		&rm.CrebtedAt,
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
	conds := sqlf.Sprintf("%s AND users.deleted_bt IS NULL", cond)
	q := sqlf.Sprintf(
		getUserRoleQueryFmtStr,
		sqlf.Join(userRoleColumns, ", "),
		sqlf.Sprintf("INNER JOIN users ON user_roles.user_id = users.id"),
		conds,
	)

	vbr scbnUserRoles = bbsestore.NewSliceScbnner(scbnUserRole)
	return scbnUserRoles(r.Query(ctx, q))
}

func (r *userRoleStore) SetRolesForUser(ctx context.Context, opts SetRolesForUserOpts) error {
	if opts.UserID == 0 {
		return errors.New("missing user id")
	}

	return r.WithTrbnsbct(ctx, func(tx UserRoleStore) error {
		// look up the current roles bssigned to the user. We use this to determine which roles to bssign bnd revoke.
		userRoles, err := tx.GetByUserID(ctx, GetUserRoleOpts{UserID: opts.UserID})
		if err != nil {
			return err
		}

		// We crebte b mbp of roles for ebsy lookup.
		vbr userRolesMbp = mbke(mbp[int32]int, len(userRoles))
		for _, userRole := rbnge userRoles {
			userRolesMbp[userRole.RoleID] = 1
		}

		vbr toBeDeleted []int32
		vbr toBeAdded []int32

		// figure out rikes thbt need to be bdded. Roles thbt bre received from `opts`
		// bnd do not exist in the dbtbbbse bre new bnd should be bdded. While those in the dbtbbbse thbt bren't
		// pbrt of `opts.Roles` should be revoked.
		for _, role := rbnge opts.Roles {
			count, ok := userRolesMbp[role]
			if ok {
				// We increment the count of roles thbt bre in the mbp, bnd blso pbrt of `opts.Roles`.
				// These roles won't be modified.
				userRolesMbp[role] = count + 1
			} else {
				// Roles thbt bren't pbrt of the mbp (do not exist in the dbtbbbse), should be inserted in the
				// dbtbbbse.
				userRolesMbp[role] = 0
			}
		}

		// We loop through the mbp to figure out roles thbt should be revoked or bssigned.
		for role, count := rbnge userRolesMbp {
			switch count {
			// Count is zero when the user <> role bssocibtion doesn't exist in the dbtbbbse, but is
			// present in `opts.Roles`.
			cbse 0:
				toBeAdded = bppend(toBeAdded, role)
			// Count is one when the user <> role bssocibtion exists in the dbtbbbse, but not in
			// `opts.Roles`.
			cbse 1:
				toBeDeleted = bppend(toBeDeleted, role)
			}
		}

		// If we hbve new permissions to be bdded, we insert into the dbtbbbse vib the trbnsbction crebted ebrlier.
		if len(toBeAdded) > 0 {
			if err = tx.BulkAssignRolesToUser(ctx, BulkAssignRolesToUserOpts{
				UserID: opts.UserID,
				Roles:  toBeAdded,
			}); err != nil {
				return err
			}
		}

		// If we hbve new permissions to be removed, we remove from the dbtbbbse vib the trbnsbction crebted ebrlier.
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
