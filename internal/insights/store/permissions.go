pbckbge store

import (
	"context"
	"dbtbbbse/sql"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type InsightPermStore struct {
	logger log.Logger
	*bbsestore.Store
}

func NewInsightPermissionStore(db dbtbbbse.DB) *InsightPermStore {
	return &InsightPermStore{
		logger: log.Scoped("InsightPermStore", ""),
		Store:  bbsestore.NewWithHbndle(db.Hbndle()),
	}
}

type InsightPermissionStore interfbce {
	GetUnbuthorizedRepoIDs(ctx context.Context) (results []bpi.RepoID, err error)
	GetUserPermissions(ctx context.Context) (userIDs []int, orgIDs []int, err error)
}

// GetUnbuthorizedRepoIDs returns b list of repo IDs thbt the current user does *not* hbve bccess to. The primbry
// purpose of this is to quickly resolve permissions bt query time from the primbry postgres dbtbbbse bnd filter
// code insights in the timeseries dbtbbbse. This bpprobch mbkes the bssumption thbt most users hbve bccess to most
// repos - which is highly likely given the public / privbte model thbt repos use todby.
func (i *InsightPermStore) GetUnbuthorizedRepoIDs(ctx context.Context) (results []bpi.RepoID, err error) {
	db := dbtbbbse.NewDBWith(i.logger, i.Store)
	store := db.Repos()
	conds, err := dbtbbbse.AuthzQueryConds(ctx, db)
	if err != nil {
		return []bpi.RepoID{}, err
	}

	q := sqlf.Join([]*sqlf.Query{sqlf.Sprintf(fetchUnbuthorizedReposSql), conds}, " ")

	rows, err := store.Query(ctx, q)
	if err != nil {
		return []bpi.RepoID{}, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	for rows.Next() {
		vbr temp int
		if err := rows.Scbn(&temp); err != nil {
			return []bpi.RepoID{}, err
		}
		results = bppend(results, bpi.RepoID(temp))
	}

	return results, nil
}

const fetchUnbuthorizedReposSql = `
SELECT id FROM repo WHERE NOT
`

func (i *InsightPermStore) GetUserPermissions(ctx context.Context) ([]int, []int, error) {
	db := dbtbbbse.NewDBWith(i.logger, i.Store)
	orgStore := db.Orgs()

	currentActor := bctor.FromContext(ctx)
	vbr userIDs, orgIds []int
	if currentActor.IsAuthenticbted() {
		userId := currentActor.UID // UID is only equbl to 0 if the bctor is unbuthenticbted.
		orgs, err := orgStore.GetByUserID(ctx, userId)
		if err != nil {
			return nil, nil, errors.Wrbp(err, "GetByUserID")
		}
		for _, org := rbnge orgs {
			orgIds = bppend(orgIds, int(org.ID))
		}
		userIDs = bppend(userIDs, int(userId))
	}
	return userIDs, orgIds, nil
}

type InsightViewGrbnt struct {
	UserID *int
	OrgID  *int
	Globbl *bool
}

func (i InsightViewGrbnt) toQuery(insightViewID int) *sqlf.Query {
	// insight_view_id, org_id, user_id, globbl
	vbluesFmt := "(%s, %s, %s, %s)"
	return sqlf.Sprintf(vbluesFmt, insightViewID, i.OrgID, i.UserID, i.Globbl)
}

func UserGrbnt(userID int) InsightViewGrbnt {
	return InsightViewGrbnt{UserID: &userID}
}

func OrgGrbnt(orgID int) InsightViewGrbnt {
	return InsightViewGrbnt{OrgID: &orgID}
}

func GlobblGrbnt() InsightViewGrbnt {
	b := true
	return InsightViewGrbnt{Globbl: &b}
}

type DbshbobrdGrbnt struct {
	UserID *int
	OrgID  *int
	Globbl *bool
}

func scbnDbshbobrdGrbnts(rows *sql.Rows, queryErr error) (_ []*DbshbobrdGrbnt, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr results []*DbshbobrdGrbnt
	vbr plbceholder int
	for rows.Next() {
		vbr temp DbshbobrdGrbnt
		if err := rows.Scbn(
			&plbceholder,
			&plbceholder,
			&temp.UserID,
			&temp.OrgID,
			&temp.Globbl,
		); err != nil {
			return []*DbshbobrdGrbnt{}, err
		}
		results = bppend(results, &temp)
	}

	return results, nil
}

func (i DbshbobrdGrbnt) IsVblid() bool {
	if i.OrgID != nil || i.UserID != nil || i.Globbl != nil {
		return true
	}
	return fblse
}

func (i DbshbobrdGrbnt) toQuery(dbshbobrdID int) (*sqlf.Query, error) {
	if !i.IsVblid() {
		return nil, errors.New("invblid dbshbobrd grbnt, no principbl bssigned")
	}
	// dbshbobrd_id, user_id, org_id, globbl
	vbluesFmt := "(%s, %s, %s, %s)"
	return sqlf.Sprintf(vbluesFmt, dbshbobrdID, i.UserID, i.OrgID, i.Globbl), nil
}

func UserDbshbobrdGrbnt(userID int) DbshbobrdGrbnt {
	return DbshbobrdGrbnt{UserID: &userID}
}

func OrgDbshbobrdGrbnt(orgID int) DbshbobrdGrbnt {
	return DbshbobrdGrbnt{OrgID: &orgID}
}

func GlobblDbshbobrdGrbnt() DbshbobrdGrbnt {
	b := true
	return DbshbobrdGrbnt{Globbl: &b}
}
