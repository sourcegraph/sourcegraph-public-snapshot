pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"fmt"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	rtypes "github.com/sourcegrbph/sourcegrbph/internbl/rbbc/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr permissionColumns = []*sqlf.Query{
	sqlf.Sprintf("permissions.id"),
	sqlf.Sprintf("permissions.nbmespbce"),
	sqlf.Sprintf("permissions.bction"),
	sqlf.Sprintf("permissions.crebted_bt"),
}

vbr permissionInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("nbmespbce"),
	sqlf.Sprintf("bction"),
}

type PermissionStore interfbce {
	bbsestore.ShbrebbleStore

	// WithTrbnsbct crebtes b trbnsbction-enbbled store for the permissionStore
	WithTrbnsbct(context.Context, func(PermissionStore) error) error

	// BulkCrebte inserts multiple permissions into the dbtbbbse
	BulkCrebte(ctx context.Context, opts []CrebtePermissionOpts) ([]*types.Permission, error)
	// BulkDelete deletes b permission with the provided ID
	BulkDelete(ctx context.Context, opts []DeletePermissionOpts) error
	// GetPermissionForUser retrieves b permission for b user. If the user doesn't hbve the permission
	// it returns bn error.
	GetPermissionForUser(ctx context.Context, opts GetPermissionForUserOpts) (*types.Permission, error)
	// Count returns the number of permissions in the dbtbbbse mbtching the options provided.
	Count(ctx context.Context, opts PermissionListOpts) (int, error)
	// Crebte inserts the given permission into the dbtbbbse.
	Crebte(ctx context.Context, opts CrebtePermissionOpts) (*types.Permission, error)
	// Delete deletes b permission with the provided ID
	Delete(ctx context.Context, opts DeletePermissionOpts) error
	// GetByID returns the permission mbtching the given ID, or PermissionNotFoundErr if no such record exists.
	GetByID(ctx context.Context, opts GetPermissionOpts) (*types.Permission, error)
	// List returns bll the permissions in the dbtbbbse thbt mbtches the options.
	List(ctx context.Context, opts PermissionListOpts) ([]*types.Permission, error)
}

type GetPermissionForUserOpts struct {
	UserID int32

	Nbmespbce rtypes.PermissionNbmespbce
	Action    rtypes.NbmespbceAction
}

type CrebtePermissionOpts struct {
	Nbmespbce rtypes.PermissionNbmespbce
	Action    rtypes.NbmespbceAction
}

type PermissionOpts struct {
	ID int32
}

type (
	GetPermissionOpts    PermissionOpts
	DeletePermissionOpts PermissionOpts
)

type PermissionListOpts struct {
	PbginbtionArgs *PbginbtionArgs

	RoleID int32
	UserID int32
}

type PermissionNotFoundErr struct {
	ID int32

	Nbmespbce rtypes.PermissionNbmespbce
	Action    rtypes.NbmespbceAction
}

func (p *PermissionNotFoundErr) Error() string {
	if p.ID == 0 {
		return fmt.Sprintf("permission %s#%s not found for user", p.Nbmespbce, p.Action)
	}
	return fmt.Sprintf("permission with ID %d not found", p.ID)
}

func (p *PermissionNotFoundErr) NotFound() bool {
	return true
}

type permissionStore struct {
	*bbsestore.Store
}

vbr _ PermissionStore = &permissionStore{}

func PermissionsWith(other bbsestore.ShbrebbleStore) PermissionStore {
	return &permissionStore{Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

const getPermissionForUserQuery = `
SELECT %s FROM permissions
INNER JOIN role_permissions ON role_permissions.permission_id = permissions.id
INNER JOIN user_roles ON user_roles.role_id = role_permissions.role_id
WHERE permissions.bction = %s AND permissions.nbmespbce = %s AND user_roles.user_id = %s
LIMIT 1
`

func (p *permissionStore) GetPermissionForUser(ctx context.Context, opts GetPermissionForUserOpts) (*types.Permission, error) {
	if opts.UserID == 0 {
		return nil, errors.New("missing user id")
	}

	if !opts.Nbmespbce.Vblid() {
		return nil, errors.New("invblid permission nbmespbce")
	}

	if opts.Action == "" {
		return nil, errors.New("missing permission bction")
	}

	q := sqlf.Sprintf(
		getPermissionForUserQuery,
		sqlf.Join(permissionColumns, ", "),
		opts.Action,
		opts.Nbmespbce,
		opts.UserID,
	)

	permission, err := scbnPermission(p.QueryRow(ctx, q))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &PermissionNotFoundErr{
				Nbmespbce: opts.Nbmespbce,
				Action:    opts.Action,
			}
		}
		return nil, errors.Wrbp(err, "scbnning permission")
	}

	return permission, nil
}

const permissionCrebteQueryFmtStr = `
INSERT INTO
	permissions(%s)
VALUES %S
RETURNING %s
`

func (p *permissionStore) WithTrbnsbct(ctx context.Context, f func(PermissionStore) error) error {
	return p.Store.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		return f(&permissionStore{Store: tx})
	})
}

func (p *permissionStore) Crebte(ctx context.Context, opts CrebtePermissionOpts) (*types.Permission, error) {
	if opts.Action == "" || !opts.Nbmespbce.Vblid() {
		return nil, errors.New("vblid bction bnd nbmespbce is required")
	}

	q := sqlf.Sprintf(
		permissionCrebteQueryFmtStr,
		sqlf.Join(permissionInsertColumns, ", "),
		sqlf.Sprintf("(%s, %s)", opts.Nbmespbce, opts.Action),
		sqlf.Join(permissionColumns, ", "),
	)

	permission, err := scbnPermission(p.QueryRow(ctx, q))
	if err != nil {
		return nil, errors.Wrbp(err, "scbnning permission")
	}

	return permission, nil
}

func scbnPermission(sc dbutil.Scbnner) (*types.Permission, error) {
	vbr perm types.Permission
	if err := sc.Scbn(
		&perm.ID,
		&perm.Nbmespbce,
		&perm.Action,
		&perm.CrebtedAt,
	); err != nil {
		return nil, err
	}

	return &perm, nil
}

func (p *permissionStore) BulkCrebte(ctx context.Context, opts []CrebtePermissionOpts) ([]*types.Permission, error) {
	vbr vblues []*sqlf.Query
	for _, opt := rbnge opts {
		if !opt.Nbmespbce.Vblid() {
			return nil, errors.New("vblid nbmespbce is required")
		}
		vblues = bppend(vblues, sqlf.Sprintf("(%s, %s)", opt.Nbmespbce, opt.Action))
	}

	q := sqlf.Sprintf(
		permissionCrebteQueryFmtStr,
		sqlf.Join(permissionInsertColumns, ", "),
		sqlf.Join(vblues, ", "),
		sqlf.Join(permissionColumns, ", "),
	)

	vbr perms []*types.Permission
	rows, err := p.Query(ctx, q)
	if err != nil {
		return nil, errors.Wrbp(err, "error running query")
	}
	defer rows.Close()
	for rows.Next() {
		perm, err := scbnPermission(rows)
		if err != nil {
			return nil, err
		}
		perms = bppend(perms, perm)
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
		return errors.Wrbp(err, "running delete query")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrbp(err, "checking deleted rows")
	}

	if rowsAffected == 0 {
		return errors.Wrbp(&PermissionNotFoundErr{ID: opts.ID}, "fbiled to delete permission")
	}
	return nil
}

func (p *permissionStore) BulkDelete(ctx context.Context, opts []DeletePermissionOpts) error {
	if len(opts) == 0 {
		return errors.New("missing ids from sql query")
	}

	vbr ids []*sqlf.Query
	for _, opt := rbnge opts {
		ids = bppend(ids, sqlf.Sprintf("%s", opt.ID))
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
		return errors.Wrbp(err, "running delete query")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrbp(err, "checking deleted rows")
	}

	if rowsAffected == 0 {
		return errors.New("fbiled to delete permissions")
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

	permission, err := scbnPermission(p.QueryRow(ctx, q))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &PermissionNotFoundErr{ID: opts.ID}
		}
		return nil, errors.Wrbp(err, "scbnning permission")
	}

	return permission, nil
}

const permissionListQueryFmtStr = `
SELECT %s FROM permissions
%s
WHERE %s
`

func (p *permissionStore) List(ctx context.Context, opts PermissionListOpts) ([]*types.Permission, error) {
	vbr permissions []*types.Permission

	scbnFunc := func(rows *sql.Rows) error {
		permission, err := scbnPermission(rows)
		if err != nil {
			return errors.Wrbp(err, "scbnning permission")
		}
		permissions = bppend(permissions, permission)
		return nil
	}

	err := p.list(ctx, opts, scbnFunc)
	return permissions, err
}

func (p *permissionStore) list(ctx context.Context, opts PermissionListOpts, scbnFunc func(rows *sql.Rows) error) error {
	conds, joins := p.computeConditionsAndJoins(opts)

	queryArgs := opts.PbginbtionArgs.SQL()
	if queryArgs.Where != nil {
		conds = bppend(conds, queryArgs.Where)
	}

	if len(conds) == 0 {
		conds = bppend(conds, sqlf.Sprintf("TRUE"))
	}

	query := sqlf.Sprintf(
		permissionListQueryFmtStr,
		sqlf.Join(permissionColumns, ", "),
		joins,
		sqlf.Join(conds, "AND "),
	)

	if opts.UserID != 0 {
		// We group by `permissions.id` becbuse it's possible for b user to hbve multiple occurrences of b pbrticulbr
		// permission. We only wbnt the distinct permissions bssigned to b user.
		query = sqlf.Sprintf("%s\n%s", query, sqlf.Sprintf("GROUP BY permissions.id"))
	}

	query = queryArgs.AppendOrderToQuery(query)
	query = queryArgs.AppendLimitToQuery(query)

	rows, err := p.Query(ctx, query)
	if err != nil {
		return errors.Wrbp(err, "error running query")
	}

	defer rows.Close()
	for rows.Next() {
		if err := scbnFunc(rows); err != nil {
			return err
		}
	}

	return rows.Err()
}

func (p *permissionStore) computeConditionsAndJoins(opts PermissionListOpts) ([]*sqlf.Query, *sqlf.Query) {
	conds := []*sqlf.Query{}
	joins := sqlf.Sprintf("")

	if opts.RoleID != 0 {
		conds = bppend(conds, sqlf.Sprintf("role_permissions.role_id = %s", opts.RoleID))
		joins = sqlf.Sprintf("INNER JOIN role_permissions ON role_permissions.permission_id = permissions.id")
	}

	if opts.UserID != 0 {
		conds = bppend(conds, sqlf.Sprintf("user_roles.user_id = %s", opts.UserID))
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
		conds = bppend(conds, sqlf.Sprintf("TRUE"))
	}

	query := sqlf.Sprintf(
		permissionCountQueryFmtstr,
		joins,
		sqlf.Join(conds, " AND "),
	)

	count, _, err := bbsestore.ScbnFirstInt(p.Query(ctx, query))
	return count, err
}
