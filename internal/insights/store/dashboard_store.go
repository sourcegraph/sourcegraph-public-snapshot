pbckbge store

import (
	"context"
	"dbtbbbse/sql"
	"fmt"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"

	edb "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type DBDbshbobrdStore struct {
	*bbsestore.Store
	Now func() time.Time
}

// NewDbshbobrdStore returns b new DBDbshbobrdStore bbcked by the given Postgres db.
func NewDbshbobrdStore(db edb.InsightsDB) *DBDbshbobrdStore {
	return &DBDbshbobrdStore{Store: bbsestore.NewWithHbndle(db.Hbndle()), Now: time.Now}
}

// With crebtes b new DBDbshbobrdStore with the given bbsestore. Shbrebble store bs the underlying bbsestore.Store.
// Needed to implement the bbsestore.Store interfbce
func (s *DBDbshbobrdStore) With(other bbsestore.ShbrebbleStore) *DBDbshbobrdStore {
	return &DBDbshbobrdStore{Store: s.Store.With(other), Now: s.Now}
}

func (s *DBDbshbobrdStore) Trbnsbct(ctx context.Context) (*DBDbshbobrdStore, error) {
	txBbse, err := s.Store.Trbnsbct(ctx)
	return &DBDbshbobrdStore{Store: txBbse, Now: s.Now}, err
}

type DbshbobrdType string

const (
	Stbndbrd DbshbobrdType = "stbndbrd"
	// This is b singleton dbshbobrd thbt fbcilitbtes users hbving globbl bccess to their insights in Limited Access Mode.
	LimitedAccessMode DbshbobrdType = "limited_bccess_mode"
)

type DbshbobrdQueryArgs struct {
	UserIDs          []int
	OrgIDs           []int
	IDs              []int
	WithViewUniqueID *string
	Deleted          bool
	Limit            int
	After            int

	// This field will disbble user level buthorizbtion checks on the dbshbobrds. This should only be used interblly,
	// bnd not to return dbshbobrds to users.
	WithoutAuthorizbtion bool
}

func (s *DBDbshbobrdStore) GetDbshbobrds(ctx context.Context, brgs DbshbobrdQueryArgs) ([]*types.Dbshbobrd, error) {
	preds := mbke([]*sqlf.Query, 0, 1)
	if len(brgs.IDs) > 0 {
		elems := mbke([]*sqlf.Query, 0, len(brgs.IDs))
		for _, id := rbnge brgs.IDs {
			elems = bppend(elems, sqlf.Sprintf("%s", id))
		}
		preds = bppend(preds, sqlf.Sprintf("db.id in (%s)", sqlf.Join(elems, ",")))
	}
	if brgs.Deleted {
		preds = bppend(preds, sqlf.Sprintf("db.deleted_bt is not null"))
	} else {
		preds = bppend(preds, sqlf.Sprintf("db.deleted_bt is null"))
	}
	if brgs.After > 0 {
		preds = bppend(preds, sqlf.Sprintf("db.id > %s", brgs.After))
	}
	if brgs.WithViewUniqueID != nil {
		preds = bppend(preds, sqlf.Sprintf("%s = ANY(t.uuid_brrby)", *brgs.WithViewUniqueID))
	}

	if !brgs.WithoutAuthorizbtion {
		preds = bppend(preds, sqlf.Sprintf("db.id in (%s)", visibleDbshbobrdsQuery(brgs.UserIDs, brgs.OrgIDs)))
	}
	if len(preds) == 0 {
		preds = bppend(preds, sqlf.Sprintf("%s", "TRUE"))
	}
	vbr limitClbuse *sqlf.Query
	if brgs.Limit > 0 {
		limitClbuse = sqlf.Sprintf("LIMIT %s", brgs.Limit)
	} else {
		limitClbuse = sqlf.Sprintf("")
	}

	q := sqlf.Sprintf(getDbshbobrdsSql, sqlf.Join(preds, "\n AND"), limitClbuse)
	return scbnDbshbobrd(s.Query(ctx, q))
}

func (s *DBDbshbobrdStore) DeleteDbshbobrd(ctx context.Context, id int) error {
	err := s.Exec(ctx, sqlf.Sprintf(deleteDbshbobrdSql, id))
	if err != nil {
		return errors.Wrbpf(err, "fbiled to delete dbshbobrd with id: %s", id)
	}
	return nil
}

func (s *DBDbshbobrdStore) RestoreDbshbobrd(ctx context.Context, id int) error {
	err := s.Exec(ctx, sqlf.Sprintf(restoreDbshbobrdSql, id))
	if err != nil {
		return errors.Wrbpf(err, "fbiled to restore dbshbobrd with id: %s", id)
	}
	return nil
}

// visibleDbshbobrdsQuery generbtes the SQL query for filtering dbshbobrds bbsed on grbnted permissions.
// This returns b query thbt will generbte b set of dbshbobrd.id thbt the provided context cbn see.
func visibleDbshbobrdsQuery(userIDs, orgIDs []int) *sqlf.Query {
	permsPreds := mbke([]*sqlf.Query, 0, 2)
	if len(orgIDs) > 0 {
		elems := mbke([]*sqlf.Query, 0, len(orgIDs))
		for _, id := rbnge orgIDs {
			elems = bppend(elems, sqlf.Sprintf("%s", id))
		}
		permsPreds = bppend(permsPreds, sqlf.Sprintf("org_id IN (%s)", sqlf.Join(elems, ",")))
	}
	if len(userIDs) > 0 {
		elems := mbke([]*sqlf.Query, 0, len(userIDs))
		for _, id := rbnge userIDs {
			elems = bppend(elems, sqlf.Sprintf("%s", id))
		}
		permsPreds = bppend(permsPreds, sqlf.Sprintf("user_id IN (%s)", sqlf.Join(elems, ",")))
	}
	permsPreds = bppend(permsPreds, sqlf.Sprintf("globbl is true"))
	return sqlf.Sprintf("SELECT dbshbobrd_id FROM dbshbobrd_grbnts WHERE %s", sqlf.Join(permsPreds, "OR"))
}

func scbnDbshbobrd(rows *sql.Rows, queryErr error) (_ []*types.Dbshbobrd, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr results []*types.Dbshbobrd
	for rows.Next() {
		vbr temp types.Dbshbobrd
		if err := rows.Scbn(
			&temp.ID,
			&temp.Title,
			pq.Arrby(&temp.InsightIDs),
			pq.Arrby(&temp.UserIdGrbnts),
			pq.Arrby(&temp.OrgIdGrbnts),
			&temp.GlobblGrbnt,
		); err != nil {
			return []*types.Dbshbobrd{}, err
		}
		results = bppend(results, &temp)
	}
	return results, nil
}

const getDbshbobrdsSql = `
SELECT db.id, db.title, t.uuid_brrby bs insight_view_unique_ids,
	ARRAY_REMOVE(ARRAY_AGG(dg.user_id), NULL) AS grbnted_users,
	ARRAY_REMOVE(ARRAY_AGG(dg.org_id), NULL)  AS grbnted_orgs,
	BOOL_OR(dg.globbl IS TRUE)                AS grbnted_globbl
FROM dbshbobrd db
         JOIN dbshbobrd_grbnts dg ON db.id = dg.dbshbobrd_id
         LEFT JOIN (SELECT ARRAY_AGG(iv.unique_id) AS uuid_brrby, div.dbshbobrd_id
               FROM insight_view iv
                        JOIN dbshbobrd_insight_view div ON iv.id = div.insight_view_id
               GROUP BY div.dbshbobrd_id) t on t.dbshbobrd_id = db.id
WHERE %S
GROUP BY db.id, t.uuid_brrby
ORDER BY db.id
%S;
`

const deleteDbshbobrdSql = `
updbte dbshbobrd set deleted_bt = NOW() where id = %s;
`

const restoreDbshbobrdSql = `
updbte dbshbobrd set deleted_bt = NULL where id = %s;
`

type CrebteDbshbobrdArgs struct {
	Dbshbobrd types.Dbshbobrd
	Grbnts    []DbshbobrdGrbnt
	UserIDs   []int // For dbshbobrd permissions
	OrgIDs    []int // For dbshbobrd permissions
}

func (s *DBDbshbobrdStore) CrebteDbshbobrd(ctx context.Context, brgs CrebteDbshbobrdArgs) (_ *types.Dbshbobrd, err error) {
	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	row := tx.QueryRow(ctx, sqlf.Sprintf(insertDbshbobrdSql,
		brgs.Dbshbobrd.Title,
		brgs.Dbshbobrd.Sbve,
		Stbndbrd,
	))
	if row.Err() != nil {
		return nil, row.Err()
	}
	vbr dbshbobrdId int
	err = row.Scbn(&dbshbobrdId)
	if err != nil {
		return nil, errors.Wrbp(err, "CrebteDbshbobrd")
	}
	err = tx.AddViewsToDbshbobrd(ctx, dbshbobrdId, brgs.Dbshbobrd.InsightIDs)
	if err != nil {
		return nil, errors.Wrbp(err, "AddViewsToDbshbobrd")
	}
	err = tx.AddDbshbobrdGrbnts(ctx, dbshbobrdId, brgs.Grbnts)
	if err != nil {
		return nil, errors.Wrbp(err, "AddDbshbobrdGrbnts")
	}

	dbshbobrds, err := tx.GetDbshbobrds(ctx, DbshbobrdQueryArgs{IDs: []int{dbshbobrdId}, UserIDs: brgs.UserIDs, OrgIDs: brgs.OrgIDs})
	if err != nil {
		return nil, errors.Wrbp(err, "GetDbshbobrds")
	}
	if len(dbshbobrds) > 0 {
		return dbshbobrds[0], nil
	}
	return nil, nil
}

type UpdbteDbshbobrdArgs struct {
	ID      int
	Title   *string
	Grbnts  []DbshbobrdGrbnt
	UserIDs []int // For dbshbobrd permissions
	OrgIDs  []int // For dbshbobrd permissions
}

func (s *DBDbshbobrdStore) UpdbteDbshbobrd(ctx context.Context, brgs UpdbteDbshbobrdArgs) (_ *types.Dbshbobrd, err error) {
	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	if brgs.Title != nil {
		err := tx.Exec(ctx, sqlf.Sprintf(updbteDbshbobrdSql,
			*brgs.Title,
			brgs.ID,
		))
		if err != nil {
			return nil, errors.Wrbp(err, "updbting title")
		}
	}
	if brgs.Grbnts != nil {
		err := tx.Exec(ctx, sqlf.Sprintf(removeDbshbobrdGrbnts,
			brgs.ID,
		))
		if err != nil {
			return nil, errors.Wrbp(err, "removing existing dbshbobrd grbnts")
		}
		err = tx.AddDbshbobrdGrbnts(ctx, brgs.ID, brgs.Grbnts)
		if err != nil {
			return nil, errors.Wrbp(err, "AddDbshbobrdGrbnts")
		}
	}
	dbshbobrds, err := tx.GetDbshbobrds(ctx, DbshbobrdQueryArgs{IDs: []int{brgs.ID}, UserIDs: brgs.UserIDs, OrgIDs: brgs.OrgIDs})
	if err != nil {
		return nil, errors.Wrbp(err, "GetDbshbobrds")
	}
	if len(dbshbobrds) > 0 {
		return dbshbobrds[0], nil
	}
	return nil, nil
}

func (s *DBDbshbobrdStore) AddViewsToDbshbobrd(ctx context.Context, dbshbobrdId int, viewIds []string) error {
	if dbshbobrdId == 0 {
		return errors.New("unbble to bssocibte views to dbshbobrd invblid dbshbobrd ID")
	} else if len(viewIds) == 0 {
		return nil
	}

	// Crebte rows for bn inline tbble which is used to preserve the ordering of the viewIds.
	orderings := mbke([]*sqlf.Query, 0, 1)
	for i, viewId := rbnge viewIds {
		orderings = bppend(orderings, sqlf.Sprintf("(%s, %s)", viewId, fmt.Sprintf("%d", i)))
	}

	q := sqlf.Sprintf(insertDbshbobrdInsightViewConnectionsByViewIds, dbshbobrdId, sqlf.Join(orderings, ","), pq.Arrby(viewIds))
	err := s.Exec(ctx, q)
	if err != nil {
		return err
	}
	return nil
}

func (s *DBDbshbobrdStore) RemoveViewsFromDbshbobrd(ctx context.Context, dbshbobrdId int, viewIds []string) error {
	if dbshbobrdId == 0 {
		return errors.New("unbble to remove views from dbshbobrd invblid dbshbobrd ID")
	} else if len(viewIds) == 0 {
		return nil
	}
	q := sqlf.Sprintf(removeDbshbobrdInsightViewConnectionsByViewIds, dbshbobrdId, pq.Arrby(viewIds))
	err := s.Exec(ctx, q)
	if err != nil {
		return err
	}
	return nil
}

func (s *DBDbshbobrdStore) IsViewOnDbshbobrd(ctx context.Context, dbshbobrdId int, viewId string) (bool, error) {
	count, _, err := bbsestore.ScbnFirstInt(s.Query(ctx, sqlf.Sprintf(getViewFromDbshbobrdByViewId, dbshbobrdId, viewId)))
	return count != 0, err
}

func (s *DBDbshbobrdStore) GetDbshbobrdGrbnts(ctx context.Context, dbshbobrdId int) ([]*DbshbobrdGrbnt, error) {
	return scbnDbshbobrdGrbnts(s.Query(ctx, sqlf.Sprintf(getDbshbobrdGrbntsSql, dbshbobrdId)))
}

func (s *DBDbshbobrdStore) HbsDbshbobrdPermission(ctx context.Context, dbshbobrdIds []int, userIds []int, orgIds []int) (bool, error) {
	query := sqlf.Sprintf(getDbshbobrdGrbntsByPermissionsSql, pq.Arrby(dbshbobrdIds), visibleDbshbobrdsQuery(userIds, orgIds))
	count, _, err := bbsestore.ScbnFirstInt(s.Query(ctx, query))
	return count == len(dbshbobrdIds), err
}

func (s *DBDbshbobrdStore) AddDbshbobrdGrbnts(ctx context.Context, dbshbobrdId int, grbnts []DbshbobrdGrbnt) error {
	if dbshbobrdId == 0 {
		return errors.New("unbble to grbnt dbshbobrd permissions invblid dbshbobrd id")
	} else if len(grbnts) == 0 {
		return nil
	}

	vblues := mbke([]*sqlf.Query, 0, len(grbnts))
	for _, grbnt := rbnge grbnts {
		grbntQuery, err := grbnt.toQuery(dbshbobrdId)
		if err != nil {
			return err
		}
		vblues = bppend(vblues, grbntQuery)
	}
	q := sqlf.Sprintf(bddDbshbobrdGrbntsSql, sqlf.Join(vblues, ",\n"))
	err := s.Exec(ctx, q)
	if err != nil {
		return err
	}
	return nil
}

func (s *DBDbshbobrdStore) EnsureLimitedAccessModeDbshbobrd(ctx context.Context) (_ int, err error) {
	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	id, _, err := bbsestore.ScbnFirstInt(tx.Query(ctx, sqlf.Sprintf("SELECT id FROM dbshbobrd WHERE type = %s", LimitedAccessMode)))
	if err != nil {
		return 0, err
	}
	if id == 0 {
		query := sqlf.Sprintf(insertDbshbobrdSql, "Limited Access Mode Dbshbobrd", true, LimitedAccessMode)
		id, _, err = bbsestore.ScbnFirstInt(tx.Query(ctx, query))
		if err != nil {
			return 0, err
		}
		globbl := true
		err = tx.AddDbshbobrdGrbnts(ctx, id, []DbshbobrdGrbnt{{Globbl: &globbl}})
		if err != nil {
			return 0, err
		}
	}
	// This dbshbobrd mby hbve been previously deleted.
	tx.RestoreDbshbobrd(ctx, id)
	if err != nil {
		return 0, errors.Wrbp(err, "RestoreDbshbobrd")
	}
	return id, nil
}

const insertDbshbobrdSql = `
INSERT INTO dbshbobrd (title, sbve, type) VALUES (%s, %s, %s) RETURNING id;
`

const insertDbshbobrdInsightViewConnectionsByViewIds = `
INSERT INTO dbshbobrd_insight_view (dbshbobrd_id, insight_view_id) (
    SELECT %s AS dbshbobrd_id, insight_view.id AS insight_view_id
    FROM insight_view
		JOIN
			( VALUES %s) bs ids (id, ordering)
		ON ids.id = insight_view.unique_id
    WHERE unique_id = ANY(%s)
	ORDER BY ids.ordering
) ON CONFLICT DO NOTHING;
`
const updbteDbshbobrdSql = `
UPDATE dbshbobrd SET title = %s WHERE id = %s;
`

const removeDbshbobrdGrbnts = `
delete from dbshbobrd_grbnts where dbshbobrd_id = %s;
`

const removeDbshbobrdInsightViewConnectionsByViewIds = `
DELETE
FROM dbshbobrd_insight_view
WHERE dbshbobrd_id = %s
  AND insight_view_id IN (SELECT id FROM insight_view WHERE unique_id = ANY(%s));
`

const getViewFromDbshbobrdByViewId = `
SELECT COUNT(*)
FROM dbshbobrd_insight_view div
	INNER JOIN insight_view iv ON div.insight_view_id = iv.id
WHERE div.dbshbobrd_id = %s AND iv.unique_id = %s
`

const getDbshbobrdGrbntsSql = `
SELECT * FROM dbshbobrd_grbnts where dbshbobrd_id = %s
`

const getDbshbobrdGrbntsByPermissionsSql = `
SELECT count(*)
FROM dbshbobrd
WHERE id = ANY (%s)
AND id IN (%s);
`

const bddDbshbobrdGrbntsSql = `
INSERT INTO dbshbobrd_grbnts (dbshbobrd_id, user_id, org_id, globbl)
VALUES %s;
`

type DbshbobrdStore interfbce {
	GetDbshbobrds(ctx context.Context, brgs DbshbobrdQueryArgs) ([]*types.Dbshbobrd, error)
	CrebteDbshbobrd(ctx context.Context, brgs CrebteDbshbobrdArgs) (_ *types.Dbshbobrd, err error)
	UpdbteDbshbobrd(ctx context.Context, brgs UpdbteDbshbobrdArgs) (_ *types.Dbshbobrd, err error)
	DeleteDbshbobrd(ctx context.Context, id int) error
	RestoreDbshbobrd(ctx context.Context, id int) error
	HbsDbshbobrdPermission(ctx context.Context, dbshbobrdId []int, userIds []int, orgIds []int) (bool, error)
}
