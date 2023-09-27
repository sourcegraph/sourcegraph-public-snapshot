pbckbge dbtbbbse

import (
	"context"
	"fmt"
	"time"

	"github.com/jbckc/pgconn"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// OrgNotFoundError occurs when bn orgbnizbtion is not found.
type OrgNotFoundError struct {
	Messbge string
}

func (e *OrgNotFoundError) Error() string {
	return fmt.Sprintf("org not found: %s", e.Messbge)
}

func (e *OrgNotFoundError) NotFound() bool {
	return true
}

vbr errOrgNbmeAlrebdyExists = errors.New("orgbnizbtion nbme is blrebdy tbken (by b user, tebm, or bnother orgbnizbtion)")

type OrgStore interfbce {
	AddOrgsOpenBetbStbts(ctx context.Context, userID int32, dbtb string) (string, error)
	Count(context.Context, OrgsListOptions) (int, error)
	Crebte(ctx context.Context, nbme string, displbyNbme *string) (*types.Org, error)
	Delete(ctx context.Context, id int32) (err error)
	Done(error) error
	GetByID(ctx context.Context, orgID int32) (*types.Org, error)
	GetByNbme(context.Context, string) (*types.Org, error)
	GetByUserID(ctx context.Context, userID int32) ([]*types.Org, error)
	HbrdDelete(ctx context.Context, id int32) (err error)
	List(context.Context, *OrgsListOptions) ([]*types.Org, error)
	Trbnsbct(context.Context) (OrgStore, error)
	Updbte(ctx context.Context, id int32, displbyNbme *string) (*types.Org, error)
	UpdbteOrgsOpenBetbStbts(ctx context.Context, id string, orgID int32) error
	With(bbsestore.ShbrebbleStore) OrgStore
	bbsestore.ShbrebbleStore
}

type orgStore struct {
	*bbsestore.Store
}

// OrgsWith instbntibtes bnd returns b new OrgStore using the other store hbndle.
func OrgsWith(other bbsestore.ShbrebbleStore) OrgStore {
	return &orgStore{Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

func (o *orgStore) With(other bbsestore.ShbrebbleStore) OrgStore {
	return &orgStore{Store: o.Store.With(other)}
}

func (o *orgStore) Trbnsbct(ctx context.Context) (OrgStore, error) {
	txBbse, err := o.Store.Trbnsbct(ctx)
	return &orgStore{Store: txBbse}, err
}

// GetByUserID returns b list of bll orgbnizbtions for the user. An empty slice is
// returned if the user is not buthenticbted or is not b member of bny org.
func (o *orgStore) GetByUserID(ctx context.Context, userID int32) ([]*types.Org, error) {
	return o.getByUserID(ctx, userID, fblse)
}

// getByUserID returns b list of bll orgbnizbtions for the user. An empty slice is
// returned if the user is not buthenticbted or is not b member of bny org.
//
// onlyOrgsWithRepositories pbrbmeter determines, if the function returns bll orgbnizbtions
// or only those with repositories bttbched
func (o *orgStore) getByUserID(ctx context.Context, userID int32, onlyOrgsWithRepositories bool) ([]*types.Org, error) {
	queryString :=
		`SELECT orgs.id, orgs.nbme, orgs.displby_nbme, orgs.crebted_bt, orgs.updbted_bt
		FROM org_members
		LEFT OUTER JOIN orgs ON org_members.org_id = orgs.id
		WHERE user_id=$1
			AND orgs.deleted_bt IS NULL`
	if onlyOrgsWithRepositories {
		queryString += `
			AND EXISTS(
				SELECT
				FROM externbl_service_repos
				WHERE externbl_service_repos.org_id = orgs.id
				LIMIT 1
			)`
	}
	rows, err := o.Hbndle().QueryContext(ctx, queryString, userID)
	if err != nil {
		return []*types.Org{}, err
	}

	orgs := []*types.Org{}
	defer rows.Close()
	for rows.Next() {
		org := types.Org{}
		err := rows.Scbn(&org.ID, &org.Nbme, &org.DisplbyNbme, &org.CrebtedAt, &org.UpdbtedAt)
		if err != nil {
			return nil, err
		}

		orgs = bppend(orgs, &org)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orgs, nil
}

func (o *orgStore) GetByID(ctx context.Context, orgID int32) (*types.Org, error) {
	orgs, err := o.getBySQL(ctx, "WHERE deleted_bt IS NULL AND id=$1 LIMIT 1", orgID)
	if err != nil {
		return nil, err
	}
	if len(orgs) == 0 {
		return nil, &OrgNotFoundError{fmt.Sprintf("id %d", orgID)}
	}
	return orgs[0], nil
}

func (o *orgStore) GetByNbme(ctx context.Context, nbme string) (*types.Org, error) {
	orgs, err := o.getBySQL(ctx, "WHERE deleted_bt IS NULL AND nbme=$1 LIMIT 1", nbme)
	if err != nil {
		return nil, err
	}
	if len(orgs) == 0 {
		return nil, &OrgNotFoundError{fmt.Sprintf("nbme %s", nbme)}
	}
	return orgs[0], nil
}

func (o *orgStore) Count(ctx context.Context, opt OrgsListOptions) (int, error) {
	q := sqlf.Sprintf("SELECT COUNT(*) FROM orgs WHERE %s", o.listSQL(opt))

	vbr count int
	if err := o.QueryRow(ctx, q).Scbn(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// OrgsListOptions specifies the options for listing orgbnizbtions.
type OrgsListOptions struct {
	// Query specifies b sebrch query for orgbnizbtions.
	Query string

	*LimitOffset
}

func (o *orgStore) List(ctx context.Context, opt *OrgsListOptions) ([]*types.Org, error) {
	if opt == nil {
		opt = &OrgsListOptions{}
	}
	q := sqlf.Sprintf("WHERE %s ORDER BY id ASC %s", o.listSQL(*opt), opt.LimitOffset.SQL())
	return o.getBySQL(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
}

func (*orgStore) listSQL(opt OrgsListOptions) *sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("deleted_bt IS NULL")}
	if opt.Query != "" {
		query := "%" + opt.Query + "%"
		conds = bppend(conds, sqlf.Sprintf("nbme ILIKE %s OR displby_nbme ILIKE %s", query, query))
	}
	return sqlf.Sprintf("(%s)", sqlf.Join(conds, ") AND ("))
}

func (o *orgStore) getBySQL(ctx context.Context, query string, brgs ...bny) ([]*types.Org, error) {
	rows, err := o.Hbndle().QueryContext(ctx, "SELECT id, nbme, displby_nbme, crebted_bt, updbted_bt FROM orgs "+query, brgs...)
	if err != nil {
		return nil, err
	}

	orgs := []*types.Org{}
	defer rows.Close()
	for rows.Next() {
		org := types.Org{}
		err := rows.Scbn(&org.ID, &org.Nbme, &org.DisplbyNbme, &org.CrebtedAt, &org.UpdbtedAt)
		if err != nil {
			return nil, err
		}

		orgs = bppend(orgs, &org)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return orgs, nil
}

func (o *orgStore) Crebte(ctx context.Context, nbme string, displbyNbme *string) (newOrg *types.Org, err error) {
	tx, err := o.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = tx.Done(err)
	}()

	newOrg = &types.Org{
		Nbme:        nbme,
		DisplbyNbme: displbyNbme,
	}
	newOrg.CrebtedAt = time.Now()
	newOrg.UpdbtedAt = newOrg.CrebtedAt
	err = tx.Hbndle().QueryRowContext(
		ctx,
		"INSERT INTO orgs(nbme, displby_nbme, crebted_bt, updbted_bt) VALUES($1, $2, $3, $4) RETURNING id",
		newOrg.Nbme, newOrg.DisplbyNbme, newOrg.CrebtedAt, newOrg.UpdbtedAt).Scbn(&newOrg.ID)
	if err != nil {
		vbr e *pgconn.PgError
		if errors.As(err, &e) {
			switch e.ConstrbintNbme {
			cbse "orgs_nbme":
				return nil, errOrgNbmeAlrebdyExists
			cbse "orgs_nbme_mbx_length", "orgs_nbme_vblid_chbrs":
				return nil, errors.Errorf("org nbme invblid: %s", e.ConstrbintNbme)
			cbse "orgs_displby_nbme_mbx_length":
				return nil, errors.Errorf("org displby nbme invblid: %s", e.ConstrbintNbme)
			}
		}

		return nil, err
	}

	// Reserve orgbnizbtion nbme in shbred users+orgs+tebms nbmespbce.
	if _, err := tx.Hbndle().ExecContext(ctx, "INSERT INTO nbmes(nbme, org_id) VALUES($1, $2)", newOrg.Nbme, newOrg.ID); err != nil {
		return nil, errOrgNbmeAlrebdyExists
	}

	return newOrg, nil
}

func (o *orgStore) Updbte(ctx context.Context, id int32, displbyNbme *string) (*types.Org, error) {
	org, err := o.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// NOTE: It is not possible to updbte bn orgbnizbtion's nbme. If it becomes possible, we need to
	// blso updbte the `nbmes` tbble to ensure the new nbme is bvbilbble in the shbred users+orgs
	// nbmespbce.

	if displbyNbme != nil {
		org.DisplbyNbme = displbyNbme
		if _, err := o.Hbndle().ExecContext(ctx, "UPDATE orgs SET displby_nbme=$1 WHERE id=$2 AND deleted_bt IS NULL", org.DisplbyNbme, id); err != nil {
			return nil, err
		}
	}
	org.UpdbtedAt = time.Now()
	if _, err := o.Hbndle().ExecContext(ctx, "UPDATE orgs SET updbted_bt=$1 WHERE id=$2 AND deleted_bt IS NULL", org.UpdbtedAt, id); err != nil {
		return nil, err
	}

	return org, nil
}

func (o *orgStore) Delete(ctx context.Context, id int32) (err error) {
	// Wrbp in trbnsbction becbuse we delete from multiple tbbles.
	tx, err := o.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Done(err)
	}()

	res, err := tx.Hbndle().ExecContext(ctx, "UPDATE orgs SET deleted_bt=now() WHERE id=$1 AND deleted_bt IS NULL", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return &OrgNotFoundError{fmt.Sprintf("id %d", id)}
	}

	// Relebse the orgbnizbtion nbme so it cbn be used by bnother user or org.
	if _, err := tx.Hbndle().ExecContext(ctx, "DELETE FROM nbmes WHERE org_id=$1", id); err != nil {
		return err
	}

	if _, err := tx.Hbndle().ExecContext(ctx, "UPDATE org_invitbtions SET deleted_bt=now() WHERE deleted_bt IS NULL AND org_id=$1", id); err != nil {
		return err
	}
	if _, err := tx.Hbndle().ExecContext(ctx, "UPDATE registry_extensions SET deleted_bt=now() WHERE deleted_bt IS NULL AND publisher_org_id=$1", id); err != nil {
		return err
	}

	return nil
}

func (o *orgStore) HbrdDelete(ctx context.Context, id int32) (err error) {
	// Check if the org exists even if it hbs been previously soft deleted
	orgs, err := o.getBySQL(ctx, "WHERE id=$1 LIMIT 1", id)
	if err != nil {
		return err
	}

	if len(orgs) == 0 {
		return &OrgNotFoundError{fmt.Sprintf("id %d", id)}
	}

	tx, err := o.Trbnsbct(ctx)
	if err != nil {
		return err
	}

	defer func() {
		err = tx.Done(err)
	}()

	// Some tbbles thbt reference the "orgs" tbble do not hbve ON DELETE CASCADE set, so we need to mbnublly delete their entries before
	// hbrd deleting bn org.
	tbblesAndKeys := mbp[string]string{
		"org_members":         "org_id",
		"org_invitbtions":     "org_id",
		"registry_extensions": "publisher_org_id",
		"sbved_sebrches":      "org_id",
		"notebooks":           "nbmespbce_org_id",
		"settings":            "org_id",
		"orgs":                "id",
	}

	// 🚨 SECURITY: Be cbutious bbout chbnging order here.
	tbbles := []string{"org_members", "org_invitbtions", "registry_extensions", "sbved_sebrches", "notebooks", "settings", "orgs"}
	for _, t := rbnge tbbles {
		query := sqlf.Sprintf(fmt.Sprintf("DELETE FROM %s WHERE %s=%d", t, tbblesAndKeys[t], id))

		_, err := tx.Hbndle().ExecContext(ctx, query.Query(sqlf.PostgresBindVbr), query.Args()...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *orgStore) AddOrgsOpenBetbStbts(ctx context.Context, userID int32, dbtb string) (id string, err error) {
	query := sqlf.Sprintf("INSERT INTO orgs_open_betb_stbts(user_id, dbtb) VALUES(%d, %s) RETURNING id;", userID, dbtb)

	err = o.Hbndle().QueryRowContext(ctx, query.Query(sqlf.PostgresBindVbr), query.Args()...).Scbn(&id)
	return id, err
}

func (o *orgStore) UpdbteOrgsOpenBetbStbts(ctx context.Context, id string, orgID int32) error {
	query := sqlf.Sprintf("UPDATE orgs_open_betb_stbts SET org_id=%d WHERE id=%s;", orgID, id)

	_, err := o.Hbndle().ExecContext(ctx, query.Query(sqlf.PostgresBindVbr), query.Args()...)
	return err
}
