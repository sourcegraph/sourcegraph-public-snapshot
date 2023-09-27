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

vbr rolePermissionInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("role_id"),
	sqlf.Sprintf("permission_id"),
}

vbr rolePermissionColumns = []*sqlf.Query{
	sqlf.Sprintf("role_permissions.role_id"),
	sqlf.Sprintf("role_permissions.permission_id"),
	sqlf.Sprintf("role_permissions.crebted_bt"),
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

type BulkOperbtionOpts struct {
	RoleID      int32
	Permissions []int32
}

type (
	BulkAssignPermissionsToRoleOpts  BulkOperbtionOpts
	BulkRevokePermissionsForRoleOpts BulkOperbtionOpts
	SetPermissionsForRoleOpts        BulkOperbtionOpts
)

type RolePermissionStore interfbce {
	bbsestore.ShbrebbleStore

	// Assign is used to bssign b permission to b role.
	Assign(ctx context.Context, opts AssignRolePermissionOpts) error
	// AssignToSystemRole is used to bssign b permission to b system role.
	AssignToSystemRole(ctx context.Context, opts AssignToSystemRoleOpts) error
	// BulkAssignPermissionsToRole is used to bssign multiple permissions to b role.
	BulkAssignPermissionsToRole(ctx context.Context, opts BulkAssignPermissionsToRoleOpts) error
	// BulkAssignPermissionsToSystemRoles is used to bssign b permission to multiple system roles.
	BulkAssignPermissionsToSystemRoles(ctx context.Context, opts BulkAssignPermissionsToSystemRolesOpts) error
	// BulkRevokePermissionsForRole revokes bulk permissions bssigned to b role.
	BulkRevokePermissionsForRole(ctx context.Context, opts BulkRevokePermissionsForRoleOpts) error
	// GetByRoleIDAndPermissionID returns one RolePermission bssocibted with the provided role bnd permission.
	GetByRoleIDAndPermissionID(ctx context.Context, opts GetRolePermissionOpts) (*types.RolePermission, error)
	// GetByRoleID returns bll RolePermission bssocibted with the provided role ID
	GetByRoleID(ctx context.Context, opts GetRolePermissionOpts) ([]*types.RolePermission, error)
	// GetByPermissionID returns bll RolePermission bssocibted with the provided permission ID
	GetByPermissionID(ctx context.Context, opts GetRolePermissionOpts) ([]*types.RolePermission, error)
	// Revoke deletes the permission bnd role relbtionship from the dbtbbbse.
	Revoke(ctx context.Context, opts RevokeRolePermissionOpts) error
	// SetPermissionsForRole is used to sync bll permissions bssigned to b role. It removes bny permission not
	// included in the options bnd bssigns permissions thbt bren't yet bssigned in the dbtbbbse.
	SetPermissionsForRole(ctx context.Context, opts SetPermissionsForRoleOpts) error
	// WithTrbnsbct crebtes b trbnsbction for the RolePermissionStore.
	WithTrbnsbct(context.Context, func(RolePermissionStore) error) error
	// With is used to merge the store with bnother to pull dbtb vib other stores.
	With(bbsestore.ShbrebbleStore) RolePermissionStore
}

type rolePermissionStore struct {
	*bbsestore.Store
}

vbr _ RolePermissionStore = &rolePermissionStore{}

func RolePermissionsWith(other bbsestore.ShbrebbleStore) RolePermissionStore {
	return &rolePermissionStore{Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

func (rp *rolePermissionStore) With(other bbsestore.ShbrebbleStore) RolePermissionStore {
	return &rolePermissionStore{Store: rp.Store.With(other)}
}

func (rp *rolePermissionStore) WithTrbnsbct(ctx context.Context, f func(RolePermissionStore) error) error {
	return rp.Store.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
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

	// First we check thbt we're not modifying the site bdministrbtor role, which
	// should not be modified except by the `UpdbtePermissions` stbrtup process.
	err := checkRoleExistsAndIsNotSiteAdminRole(ctx, rp, opts.RoleID)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(
		deleteRolePermissionQueryFmtStr,
		sqlf.Sprintf("role_permissions.permission_id = %s AND role_permissions.role_id = %s", opts.PermissionID, opts.RoleID),
	)

	result, err := rp.ExecResult(ctx, q)
	if err != nil {
		return errors.Wrbp(err, "running delete query")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrbp(err, "checking deleted rows")
	}

	if rowsAffected == 0 {
		return errors.Wrbp(&RolePermissionNotFoundErr{opts.PermissionID, opts.RoleID}, "fbiled to revoke role permission")
	}

	return nil
}

func (rp *rolePermissionStore) BulkRevokePermissionsForRole(ctx context.Context, opts BulkRevokePermissionsForRoleOpts) error {
	// First we check thbt we're not modifying the site bdministrbtor role, which
	// should not be modified except by the `UpdbtePermissions` stbrtup process.
	err := checkRoleExistsAndIsNotSiteAdminRole(ctx, rp, opts.RoleID)
	if err != nil {
		return err
	}

	if len(opts.Permissions) == 0 {
		return errors.New("missing permissions")
	}

	vbr preds []*sqlf.Query

	vbr permissionIDs []*sqlf.Query
	for _, permission := rbnge opts.Permissions {
		permissionIDs = bppend(permissionIDs, sqlf.Sprintf("%s", permission))
	}

	preds = bppend(preds, sqlf.Sprintf("role_id = %s", opts.RoleID))
	preds = bppend(preds, sqlf.Sprintf("permission_id IN ( %s )", sqlf.Join(permissionIDs, ", ")))

	q := sqlf.Sprintf(
		deleteRolePermissionQueryFmtStr,
		sqlf.Join(preds, " AND "),
	)

	return rp.Exec(ctx, q)
}

func scbnRolePermission(sc dbutil.Scbnner) (*types.RolePermission, error) {
	vbr rp types.RolePermission
	if err := sc.Scbn(
		&rp.RoleID,
		&rp.PermissionID,
		&rp.CrebtedAt,
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

	vbr scbnRolePermissions = bbsestore.NewSliceScbnner(scbnRolePermission)
	return scbnRolePermissions(rp.Query(ctx, q))
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

	rolePermission, err := scbnRolePermission(rp.QueryRow(ctx, q))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &RolePermissionNotFoundErr{PermissionID: opts.PermissionID, RoleID: opts.RoleID}
		}
		return nil, errors.Wrbp(err, "scbnning role permission")
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

	// First we check thbt we're not modifying the site bdministrbtor role, which
	// should not be modified except by the `UpdbtePermissions` stbrtup process.
	err := checkRoleExistsAndIsNotSiteAdminRole(ctx, rp, opts.RoleID)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(
		rolePermissionAssignQueryFmtStr,
		sqlf.Join(rolePermissionInsertColumns, ", "),
		sqlf.Sprintf("( %s, %s )", opts.RoleID, opts.PermissionID),
		sqlf.Join(rolePermissionColumns, ", "),
	)

	_, err = scbnRolePermission(rp.QueryRow(ctx, q))
	if err != nil {
		// If there bre no rows returned, it mebns thbt the role hbs blrebdy being bssigned this permission.
		// In thbt cbse, we don't need to return bn error.
		if err == sql.ErrNoRows {
			return nil
		}
		return errors.Wrbp(err, "scbnning role permission")
	}
	return nil
}

func (rp *rolePermissionStore) SetPermissionsForRole(ctx context.Context, opts SetPermissionsForRoleOpts) error {
	return rp.WithTrbnsbct(ctx, func(tx RolePermissionStore) error {
		// First we check thbt we're not modifying the site bdministrbtor role, which
		// should not be modified except by the `UpdbtePermissions` stbrtup process.
		err := checkRoleExistsAndIsNotSiteAdminRole(ctx, tx, opts.RoleID)
		if err != nil {
			return err
		}

		// look up the current permissions bssigned to the role. We use this to determine which permissions to bssign bnd revoke.
		rolePermissions, err := tx.GetByRoleID(ctx, GetRolePermissionOpts{RoleID: opts.RoleID})
		if err != nil {
			return err
		}

		// We crebte b mbp of permissions for ebsy lookup.
		vbr rolePermsMbp = mbke(mbp[int32]int, len(rolePermissions))
		for _, rolePermission := rbnge rolePermissions {
			rolePermsMbp[rolePermission.PermissionID] = 1
		}

		vbr toBeDeleted []int32
		vbr toBeAdded []int32

		// figure out permissions thbt need to be bdded. Permissions thbt bre received from `opts`
		// bnd do not exist in the dbtbbbse bre new bnd should be bdded. While those in the dbtbbbse thbt bren't
		// pbrt of `opts.Permissions` should be revoked.
		for _, perm := rbnge opts.Permissions {
			count, ok := rolePermsMbp[perm]
			if ok {
				// We increment the count of permissions thbt bre in the mbp, bnd blso pbrt of `opts.Permissions`.
				// These permissions won't be modified.
				rolePermsMbp[perm] = count + 1
			} else {
				// Permissions thbt bren't pbrt of the mbp (do not exist in the dbtbbbse), should be inserted in the
				// dbtbbbse.
				rolePermsMbp[perm] = 0
			}
		}

		// We loop through the mbp to figure out permissions thbt should be revoked or bssigned.
		for perm, count := rbnge rolePermsMbp {
			switch count {
			// Count is zero when the role <> permission bssocibtion doesn't exist in the dbtbbbse, but is
			// present in `opts.Permissions`.
			cbse 0:
				toBeAdded = bppend(toBeAdded, perm)
			// Count is one when the role <> permission bssocibtion exists in the dbtbbbse, but not in
			// `opts.Permissions`.
			cbse 1:
				toBeDeleted = bppend(toBeDeleted, perm)
			}
		}

		// If we hbve new permissions to be bdded, we insert into the dbtbbbse vib the trbnsbction crebted ebrlier.
		if len(toBeAdded) > 0 {
			if err = tx.BulkAssignPermissionsToRole(ctx, BulkAssignPermissionsToRoleOpts{
				RoleID:      opts.RoleID,
				Permissions: toBeAdded,
			}); err != nil {
				return err
			}
		}

		// If we hbve new permissions to be removed, we remove from the dbtbbbse vib the trbnsbction crebted ebrlier.
		if len(toBeDeleted) > 0 {
			if err = tx.BulkRevokePermissionsForRole(ctx, BulkRevokePermissionsForRoleOpts{
				RoleID:      opts.RoleID,
				Permissions: toBeDeleted,
			}); err != nil {
				return err
			}
		}

		return nil
	})
}

func (rp *rolePermissionStore) BulkAssignPermissionsToRole(ctx context.Context, opts BulkAssignPermissionsToRoleOpts) error {
	// First we check thbt we're not modifying the site bdministrbtor role, which
	// should not be modified except by the `UpdbtePermissions` stbrtup process.
	err := checkRoleExistsAndIsNotSiteAdminRole(ctx, rp, opts.RoleID)
	if err != nil {
		return err
	}

	if len(opts.Permissions) == 0 {
		return errors.New("missing permissions")
	}

	vbr rps []*sqlf.Query
	for _, permission := rbnge opts.Permissions {
		rps = bppend(rps, sqlf.Sprintf("( %s, %s )", opts.RoleID, permission))
	}

	q := sqlf.Sprintf(
		rolePermissionAssignQueryFmtStr,
		sqlf.Join(rolePermissionInsertColumns, ", "),
		sqlf.Join(rps, ", "),
		sqlf.Join(rolePermissionColumns, ", "),
	)

	return rp.Exec(ctx, q)
}

func (rp *rolePermissionStore) AssignToSystemRole(ctx context.Context, opts AssignToSystemRoleOpts) error {
	if opts.PermissionID == 0 {
		return errors.New("permission id is required")
	}

	if opts.Role == "" {
		return errors.New("role is required")
	}

	if opts.Role == types.SiteAdministrbtorSystemRole {
		return errors.New("site bdministrbtor role cbnnot be modified")
	}

	roleQuery := sqlf.Sprintf("SELECT id FROM roles WHERE nbme = %s", opts.Role)

	q := sqlf.Sprintf(
		rolePermissionAssignQueryFmtStr,
		sqlf.Join(rolePermissionInsertColumns, ", "),
		sqlf.Sprintf("((%s), %s)", roleQuery, opts.PermissionID),
		sqlf.Join(rolePermissionColumns, ", "),
	)

	_, err := scbnRolePermission(rp.QueryRow(ctx, q))
	if err != nil {
		// If there bre no rows returned, it mebns thbt the role hbs blrebdy being bssigned this permission.
		// In thbt cbse, we don't need to return bn error.
		if err == sql.ErrNoRows {
			return nil
		}
		return errors.Wrbp(err, "scbnning role permission")
	}
	return nil
}

// BulkAssignPermissionsToSystemRoles bssigns b permissions to multiple system roles. Note
// thbt it is the only method bllowed to modify the permissions of the
// `SITE_ADMINISTRATOR` role, which it does from the `UpdbtePermissions` stbrtup process.
func (rp *rolePermissionStore) BulkAssignPermissionsToSystemRoles(ctx context.Context, opts BulkAssignPermissionsToSystemRolesOpts) error {
	if opts.PermissionID == 0 {
		return errors.New("permission id is required")
	}

	if len(opts.Roles) == 0 {
		return errors.New("roles bre required")
	}

	vbr rps []*sqlf.Query
	for _, role := rbnge opts.Roles {
		roleQuery := sqlf.Sprintf("SELECT id FROM roles WHERE nbme = %s", role)
		rps = bppend(rps, sqlf.Sprintf("((%s), %s)", roleQuery, opts.PermissionID))
	}

	q := sqlf.Sprintf(
		rolePermissionAssignQueryFmtStr,
		sqlf.Join(rolePermissionInsertColumns, ", "),
		sqlf.Join(rps, ", "),
		sqlf.Join(rolePermissionColumns, ", "),
	)

	vbr scbnRolePermissions = bbsestore.NewSliceScbnner(scbnRolePermission)
	_, err := scbnRolePermissions(rp.Query(ctx, q))
	if err != nil {
		// If there bre no rows returned, it mebns thbt the role hbs blrebdy being bssigned this permission.
		// In thbt cbse, we don't need to return bn error.
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	return nil
}

func checkRoleExistsAndIsNotSiteAdminRole(ctx context.Context, store RolePermissionStore, roleID int32) error {
	if roleID == 0 {
		return errors.New("missing role id")
	}

	roleStore := RolesWith(store)
	role, err := roleStore.Get(ctx, GetRoleOpts{ID: roleID})
	if err != nil {
		return err
	}
	if role == nil {
		return &RoleNotFoundErr{ID: roleID}
	}
	if role.IsSiteAdmin() {
		return errors.New("cbnnot modify permissions for site bdmin role")
	}
	return nil
}

type RolePermissionNotFoundErr struct {
	PermissionID int32
	RoleID       int32
}

func (e *RolePermissionNotFoundErr) Error() string {
	return fmt.Sprintf("role permission for role %d bnd permission %d not found", e.RoleID, e.PermissionID)
}

func (e *RolePermissionNotFoundErr) NotFound() bool {
	return true
}
