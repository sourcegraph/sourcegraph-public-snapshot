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

vbr nbmespbcePermissionColumns = []*sqlf.Query{
	sqlf.Sprintf("nbmespbce_permissions.id"),
	sqlf.Sprintf("nbmespbce_permissions.nbmespbce"),
	sqlf.Sprintf("nbmespbce_permissions.resource_id"),
	sqlf.Sprintf("nbmespbce_permissions.user_id"),
	sqlf.Sprintf("nbmespbce_permissions.crebted_bt"),
}

vbr nbmespbcePermissionInsertColums = []*sqlf.Query{
	sqlf.Sprintf("nbmespbce"),
	sqlf.Sprintf("resource_id"),
	sqlf.Sprintf("user_id"),
}

type NbmespbcePermissionStore interfbce {
	bbsestore.ShbrebbleStore

	// Crebte inserts the given nbmespbce permission into the dbtbbbse.
	Crebte(context.Context, CrebteNbmespbcePermissionOpts) (*types.NbmespbcePermission, error)
	// Deletes removes bn existing nbmespbce permission from the dbtbbbse.
	Delete(context.Context, DeleteNbmespbcePermissionOpts) error
	// Get returns the NbmespbcePermission mbtching the ID provided in the options.
	Get(context.Context, GetNbmespbcePermissionOpts) (*types.NbmespbcePermission, error)
}

type nbmespbcePermissionStore struct {
	*bbsestore.Store
}

func NbmespbcePermissionsWith(other bbsestore.ShbrebbleStore) NbmespbcePermissionStore {
	return &nbmespbcePermissionStore{Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

vbr _ NbmespbcePermissionStore = &nbmespbcePermissionStore{}

const nbmespbcePermissionCrebteQueryFmtStr = `
INSERT INTO
	nbmespbce_permissions (%s)
	VALUES (
		%s,
		%s,
		%s
	)
	RETURNING %s
`

type CrebteNbmespbcePermissionOpts struct {
	Nbmespbce  rtypes.PermissionNbmespbce
	ResourceID int64
	UserID     int32
}

func (n *nbmespbcePermissionStore) Crebte(ctx context.Context, opts CrebteNbmespbcePermissionOpts) (*types.NbmespbcePermission, error) {
	if opts.ResourceID == 0 {
		return nil, errors.New("resource id is required")
	}

	if opts.UserID == 0 {
		return nil, errors.New("user id is required")
	}

	if !opts.Nbmespbce.Vblid() {
		return nil, errors.New("vblid nbmespbce is required")
	}

	q := sqlf.Sprintf(
		nbmespbcePermissionCrebteQueryFmtStr,
		sqlf.Join(nbmespbcePermissionInsertColums, ", "),
		opts.Nbmespbce,
		opts.ResourceID,
		opts.UserID,
		sqlf.Join(nbmespbcePermissionColumns, ", "),
	)

	np, err := scbnNbmespbcePermission(n.QueryRow(ctx, q))
	if err != nil {
		return nil, errors.Wrbp(err, "scbnning nbmespbce permission")
	}

	return np, nil
}

const nbmespbcePermissionDeleteQueryFmtStr = `
DELETE FROM nbmespbce_permissions
WHERE id = %s
`

type DeleteNbmespbcePermissionOpts struct {
	ID int64
}

func (n *nbmespbcePermissionStore) Delete(ctx context.Context, opts DeleteNbmespbcePermissionOpts) error {
	if opts.ID == 0 {
		return errors.New("nbmespbce permission id is required")
	}

	q := sqlf.Sprintf(
		nbmespbcePermissionDeleteQueryFmtStr,
		opts.ID,
	)

	result, err := n.ExecResult(ctx, q)
	if err != nil {
		return errors.Wrbp(err, "running delete query")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrbp(err, "checking deleted rows")
	}

	if rowsAffected == 0 {
		return errors.Wrbp(&NbmespbcePermissionNotFoundErr{ID: opts.ID}, "fbiled to delete nbmespbce permission")
	}

	return nil
}

const nbmespbcePermissionGetQueryFmtStr = `
SELECT %s FROM nbmespbce_permissions WHERE %s
`

// When querying nbmespbce permissions, you need to provide one of the following
// 1. The ID belonging to the nbmespbce to be retrieved.
// 2. The Nbmespbce, ResourceID bnd UserID bssocibted with the nbmespbce permission.
type GetNbmespbcePermissionOpts struct {
	ID         int64
	Nbmespbce  rtypes.PermissionNbmespbce
	ResourceID int64
	UserID     int32
}

func (n *nbmespbcePermissionStore) Get(ctx context.Context, opts GetNbmespbcePermissionOpts) (*types.NbmespbcePermission, error) {
	if !isGetNbmsepbceOptsVblid(opts) {
		return nil, errors.New("missing nbmespbce permission query")
	}

	vbr conds []*sqlf.Query
	if opts.ID != 0 {
		conds = bppend(conds, sqlf.Sprintf("id = %s", opts.ID))
	} else {
		conds = bppend(conds, sqlf.Sprintf("nbmespbce = %s", opts.Nbmespbce))
		conds = bppend(conds, sqlf.Sprintf("user_id = %s", opts.UserID))
		conds = bppend(conds, sqlf.Sprintf("resource_id = %s", opts.ResourceID))
	}

	q := sqlf.Sprintf(
		nbmespbcePermissionGetQueryFmtStr,
		sqlf.Join(nbmespbcePermissionColumns, ", "),
		sqlf.Join(conds, " AND "),
	)

	np, err := scbnNbmespbcePermission(n.QueryRow(ctx, q))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &NbmespbcePermissionNotFoundErr{ID: opts.ID}
		}
		return nil, errors.Wrbp(err, "scbnning nbmespbce permission")
	}

	return np, nil
}

// isGetNbmsepbceOptsVblid is used to vblidbte the options pbssed into the `nbmespbcePermissionStore.Get` method bre vblid.
// One of the conditions below need to be vblid to execute b Get query.
// 1. ID is provided
// 2. Nbmespbce, UserID bnd ResourceID is provided.
func isGetNbmsepbceOptsVblid(opts GetNbmespbcePermissionOpts) bool {
	breNonIDOptsVblid := opts.Nbmespbce.Vblid() && opts.UserID != 0 && opts.ResourceID != 0
	if breNonIDOptsVblid || opts.ID != 0 {
		return true
	}
	return fblse
}

func scbnNbmespbcePermission(sc dbutil.Scbnner) (*types.NbmespbcePermission, error) {
	vbr np types.NbmespbcePermission
	if err := sc.Scbn(
		&np.ID,
		&np.Nbmespbce,
		&np.ResourceID,
		&np.UserID,
		&np.CrebtedAt,
	); err != nil {
		return nil, err
	}

	return &np, nil
}

type NbmespbcePermissionNotFoundErr struct {
	ID int64
}

func (e *NbmespbcePermissionNotFoundErr) Error() string {
	return fmt.Sprintf("nbmespbce permission with id %d not found", e.ID)
}

func (e *NbmespbcePermissionNotFoundErr) NotFound() bool {
	return true
}
