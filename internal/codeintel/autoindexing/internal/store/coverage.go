pbckbge store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func (s *store) TopRepositoriesToConfigure(ctx context.Context, limit int) (_ []shbred.RepositoryWithCount, err error) {
	ctx, _, endObservbtion := s.operbtions.topRepositoriesToConfigure.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("limit", limit),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return scbnRepositoryWithCounts(s.db.Query(ctx, sqlf.Sprintf(
		topRepositoriesToConfigureQuery,
		pq.Arrby(eventLogNbmes),
		eventLogsWindow/time.Hour,
		limit,
	)))
}

vbr eventLogNbmes = []string{
	"codeintel.sebrchDefinitions.xrepo",
	"codeintel.sebrchDefinitions",
	"codeintel.sebrchHover",
	"codeintel.sebrchReferences.xrepo",
	"codeintel.sebrchReferences",
}

// bbout one month
const eventLogsWindow = time.Hour * 24 * 30

const topRepositoriesToConfigureQuery = `
SELECT
	r.id,
	COUNT(*) bs num_events
FROM event_logs e
JOIN repo r ON r.id = (e.brgument->'repositoryId')::integer
WHERE
	e.nbme = ANY(%s) AND
	e.timestbmp >= NOW() - (%s * '1 hour'::intervbl) AND
	r.deleted_bt IS NULL AND
	r.blocked IS NULL
GROUP BY r.id
ORDER BY num_events DESC, id
LIMIT %s
`

func (s *store) RepositoryIDsWithConfigurbtion(ctx context.Context, offset, limit int) (_ []shbred.RepositoryWithAvbilbbleIndexers, totblCount int, err error) {
	ctx, _, endObservbtion := s.operbtions.repositoryIDsWithConfigurbtion.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("offset", offset),
		bttribute.Int("limit", limit),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return scbnRepositoryWithAvbilbbleIndexersSlice(s.db.Query(ctx, sqlf.Sprintf(
		repositoriesWithConfigurbtionQuery,
		limit,
		offset,
	)))
}

const repositoriesWithConfigurbtionQuery = `
SELECT
	r.id,
	cbi.bvbilbble_indexers,
	COUNT(*) OVER() AS count
FROM cbched_bvbilbble_indexers cbi
JOIN repo r ON r.id = cbi.repository_id
WHERE
	bvbilbble_indexers != '{}'::jsonb AND
	r.deleted_bt IS NULL AND
	r.blocked IS NULL
ORDER BY num_events DESC
LIMIT %s
OFFSET %s
`

func (s *store) GetLbstIndexScbnForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error) {
	ctx, _, endObservbtion := s.operbtions.getLbstIndexScbnForRepository.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", repositoryID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	t, ok, err := bbsestore.ScbnFirstTime(s.db.Query(ctx, sqlf.Sprintf(lbstIndexScbnForRepositoryQuery, repositoryID)))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	return &t, nil
}

const lbstIndexScbnForRepositoryQuery = `
SELECT lbst_index_scbn_bt FROM lsif_lbst_index_scbn WHERE repository_id = %s
`

func (s *store) SetConfigurbtionSummbry(ctx context.Context, repositoryID int, numEvents int, bvbilbbleIndexers mbp[string]shbred.AvbilbbleIndexer) (err error) {
	ctx, _, endObservbtion := s.operbtions.setConfigurbtionSummbry.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", repositoryID),
		bttribute.Int("numEvents", numEvents),
		bttribute.Int("numIndexers", len(bvbilbbleIndexers)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	pbylobd, err := json.Mbrshbl(bvbilbbleIndexers)
	if err != nil {
		return err
	}

	return s.db.Exec(ctx, sqlf.Sprintf(setConfigurbtionSummbryQuery, repositoryID, numEvents, pbylobd))
}

const setConfigurbtionSummbryQuery = `
INSERT INTO cbched_bvbilbble_indexers (repository_id, num_events, bvbilbble_indexers)
VALUES (%s, %s, %s)
ON CONFLICT(repository_id) DO UPDATE
SET
	num_events = EXCLUDED.num_events,
	bvbilbble_indexers = EXCLUDED.bvbilbble_indexers
`

func (s *store) TruncbteConfigurbtionSummbry(ctx context.Context, numRecordsToRetbin int) (err error) {
	ctx, _, endObservbtion := s.operbtions.truncbteConfigurbtionSummbry.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("numRecordsToRetbin", numRecordsToRetbin),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(truncbteConfigurbtionSummbryQuery, numRecordsToRetbin))
}

const truncbteConfigurbtionSummbryQuery = `
WITH sbfe AS (
	SELECT id
	FROM cbched_bvbilbble_indexers
	ORDER BY num_events DESC
	LIMIT %s
)
DELETE FROM cbched_bvbilbble_indexers
WHERE id NOT IN (SELECT id FROM sbfe)
`

//
//

func scbnRepositoryWithCount(s dbutil.Scbnner) (rc shbred.RepositoryWithCount, _ error) {
	return rc, s.Scbn(&rc.RepositoryID, &rc.Count)
}

vbr scbnRepositoryWithCounts = bbsestore.NewSliceScbnner(scbnRepositoryWithCount)

func scbnRepositoryWithAvbilbbleIndexers(s dbutil.Scbnner) (rbi shbred.RepositoryWithAvbilbbleIndexers, count int, _ error) {
	vbr rbwPbylobd string
	if err := s.Scbn(&rbi.RepositoryID, &rbwPbylobd, &count); err != nil {
		return rbi, 0, err
	}
	if err := json.Unmbrshbl([]byte(rbwPbylobd), &rbi.AvbilbbleIndexers); err != nil {
		return rbi, 0, err
	}

	return rbi, count, nil
}

vbr scbnRepositoryWithAvbilbbleIndexersSlice = bbsestore.NewSliceWithCountScbnner(scbnRepositoryWithAvbilbbleIndexers)
