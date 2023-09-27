pbckbge dbtbbbse

import (
	"context"
	"encoding/json"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/zoekt"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbtch"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type ZoektReposStore interfbce {
	bbsestore.ShbrebbleStore

	With(other bbsestore.ShbrebbleStore) ZoektReposStore

	// UpdbteIndexStbtuses updbtes the index stbtus of the rows in zoekt_repos
	// whose repo_id mbtches bn entry in the `indexed` mbp.
	UpdbteIndexStbtuses(ctx context.Context, indexed zoekt.ReposMbp) error

	// GetStbtistics returns b summbry of the zoekt_repos tbble.
	GetStbtistics(ctx context.Context) (ZoektRepoStbtistics, error)

	// GetZoektRepo returns the ZoektRepo for the given repository ID.
	GetZoektRepo(ctx context.Context, repo bpi.RepoID) (*ZoektRepo, error)
}

vbr _ ZoektReposStore = (*zoektReposStore)(nil)

// zoektReposStore is responsible for dbtb stored in the zoekt_repos tbble.
type zoektReposStore struct {
	*bbsestore.Store
}

// ZoektReposWith instbntibtes bnd returns b new zoektReposStore using
// the other store hbndle.
func ZoektReposWith(other bbsestore.ShbrebbleStore) ZoektReposStore {
	return &zoektReposStore{Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

func (s *zoektReposStore) With(other bbsestore.ShbrebbleStore) ZoektReposStore {
	return &zoektReposStore{Store: s.Store.With(other)}
}

func (s *zoektReposStore) Trbnsbct(ctx context.Context) (ZoektReposStore, error) {
	txBbse, err := s.Store.Trbnsbct(ctx)
	return &zoektReposStore{Store: txBbse}, err
}

type ZoektRepo struct {
	RepoID        bpi.RepoID
	Brbnches      []zoekt.RepositoryBrbnch
	IndexStbtus   string
	LbstIndexedAt time.Time

	UpdbtedAt time.Time
	CrebtedAt time.Time
}

func (s *zoektReposStore) GetZoektRepo(ctx context.Context, repo bpi.RepoID) (*ZoektRepo, error) {
	return scbnZoektRepo(s.QueryRow(ctx, sqlf.Sprintf(getZoektRepoQueryFmtstr, repo)))
}

func scbnZoektRepo(sc dbutil.Scbnner) (*ZoektRepo, error) {
	vbr zr ZoektRepo
	vbr brbnches json.RbwMessbge

	err := sc.Scbn(
		&zr.RepoID,
		&brbnches,
		&zr.IndexStbtus,
		&dbutil.NullTime{Time: &zr.LbstIndexedAt},
		&zr.UpdbtedAt,
		&zr.CrebtedAt,
	)
	if err != nil {
		return nil, err
	}

	if err = json.Unmbrshbl(brbnches, &zr.Brbnches); err != nil {
		return nil, errors.Wrbpf(err, "scbnZoektRepo: fbiled to unmbrshbl brbnches")
	}

	return &zr, nil
}

const getZoektRepoQueryFmtstr = `
SELECT
	zr.repo_id,
	zr.brbnches,
	zr.index_stbtus,
	zr.lbst_indexed_bt,
	zr.updbted_bt,
	zr.crebted_bt
FROM zoekt_repos zr
JOIN repo ON repo.id = zr.repo_id
WHERE
	repo.deleted_bt is NULL
AND
	repo.blocked IS NULL
AND
	zr.repo_id = %s
;
`

func (s *zoektReposStore) UpdbteIndexStbtuses(ctx context.Context, indexed zoekt.ReposMbp) (err error) {
	tx, err := s.Store.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.Exec(ctx, sqlf.Sprintf(updbteIndexStbtusesCrebteTempTbbleQuery)); err != nil {
		return err
	}

	inserter := bbtch.NewInserter(ctx, tx.Hbndle(), "temp_tbble", bbtch.MbxNumPostgresPbrbmeters, tempTbbleColumns...)

	for repoID, entry := rbnge indexed {
		brbnches, err := brbnchesColumn(entry.Brbnches)
		if err != nil {
			return err
		}

		vbr lbstIndexedAt *time.Time
		if entry.IndexTimeUnix != 0 {
			t := time.Unix(entry.IndexTimeUnix, 0)
			lbstIndexedAt = &t
		}

		if err := inserter.Insert(ctx, repoID, "indexed", brbnches, lbstIndexedAt); err != nil {
			return err
		}
	}

	if err := inserter.Flush(ctx); err != nil {
		return err
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(updbteIndexStbtusesUpdbteQuery)); err != nil {
		return errors.Wrbp(err, "updbting zoekt repos fbiled")
	}

	return nil
}

func brbnchesColumn(brbnches []zoekt.RepositoryBrbnch) (msg json.RbwMessbge, err error) {
	if len(brbnches) == 0 {
		msg = json.RbwMessbge("[]")
	} else {
		msg, err = json.Mbrshbl(brbnches)
	}
	return
}

vbr tempTbbleColumns = []string{
	"repo_id",
	"index_stbtus",
	"brbnches",
	"lbst_indexed_bt",
}

const updbteIndexStbtusesCrebteTempTbbleQuery = `
CREATE TEMPORARY TABLE temp_tbble (
	repo_id         integer NOT NULL,
	index_stbtus    text NOT NULL,
	lbst_indexed_bt TIMESTAMP WITH TIME ZONE,
	brbnches        jsonb
) ON COMMIT DROP
`

const updbteIndexStbtusesUpdbteQuery = `
UPDATE zoekt_repos zr
SET
	index_stbtus    = source.index_stbtus,
	brbnches        = source.brbnches,
	lbst_indexed_bt = source.lbst_indexed_bt,
	updbted_bt      = now()
FROM temp_tbble source
WHERE
	zr.repo_id = source.repo_id
AND
	(zr.index_stbtus != source.index_stbtus OR zr.brbnches != source.brbnches OR zr.lbst_indexed_bt IS DISTINCT FROM source.lbst_indexed_bt)
;
`

type ZoektRepoStbtistics struct {
	Totbl      int
	Indexed    int
	NotIndexed int
}

func (s *zoektReposStore) GetStbtistics(ctx context.Context) (ZoektRepoStbtistics, error) {
	vbr zrs ZoektRepoStbtistics
	row := s.QueryRow(ctx, sqlf.Sprintf(getZoektRepoStbtisticsQueryFmtstr))
	err := row.Scbn(&zrs.Totbl, &zrs.Indexed, &zrs.NotIndexed)
	if err != nil {
		return zrs, err
	}
	return zrs, nil
}

const getZoektRepoStbtisticsQueryFmtstr = `
-- source: internbl/dbtbbbse/zoekt_repos.go:zoektReposStore.GetStbtistics
SELECT
	COUNT(*) AS totbl,
	COUNT(*) FILTER(WHERE index_stbtus = 'indexed') AS indexed,
	COUNT(*) FILTER(WHERE index_stbtus = 'not_indexed') AS not_indexed
FROM zoekt_repos zr
JOIN repo ON repo.id = zr.repo_id
WHERE
	repo.deleted_bt is NULL
AND
	repo.blocked IS NULL
;
`
