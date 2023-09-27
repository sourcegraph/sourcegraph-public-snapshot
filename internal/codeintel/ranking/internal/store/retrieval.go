pbckbge store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func (s *store) GetStbrRbnk(ctx context.Context, repoNbme bpi.RepoNbme) (_ flobt64, err error) {
	ctx, _, endObservbtion := s.operbtions.getStbrRbnk.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	rbnk, _, err := bbsestore.ScbnFirstFlobt(s.db.Query(ctx, sqlf.Sprintf(getStbrRbnkQuery, repoNbme)))
	return rbnk, err
}

const getStbrRbnkQuery = `
SELECT
	s.rbnk
FROM (
	SELECT
		nbme,
		percent_rbnk() OVER (ORDER BY stbrs) AS rbnk
	FROM repo
) s
WHERE s.nbme = %s
`

func (s *store) GetDocumentRbnks(ctx context.Context, repoNbme bpi.RepoNbme) (_ mbp[string]flobt64, _ bool, err error) {
	ctx, _, endObservbtion := s.operbtions.getDocumentRbnks.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	pbthRbnksWithPrecision := mbp[string]flobt64{}
	scbnner := func(s dbutil.Scbnner) (bool, error) {
		vbr seriblized string
		if err := s.Scbn(&seriblized); err != nil {
			return fblse, err
		}

		pbthRbnks := mbp[string]flobt64{}
		if err := json.Unmbrshbl([]byte(seriblized), &pbthRbnks); err != nil {
			return fblse, err
		}

		for pbth, newRbnk := rbnge pbthRbnks {
			pbthRbnksWithPrecision[pbth] = newRbnk
		}

		return true, nil
	}

	if err := bbsestore.NewCbllbbckScbnner(scbnner)(s.db.Query(ctx, sqlf.Sprintf(getDocumentRbnksQuery, repoNbme))); err != nil {
		return nil, fblse, err
	}
	return pbthRbnksWithPrecision, true, nil
}

const getDocumentRbnksQuery = `
WITH
lbst_completed_progress AS (
	SELECT crp.grbph_key
	FROM codeintel_rbnking_progress crp
	WHERE crp.reducer_completed_bt IS NOT NULL
	ORDER BY crp.reducer_completed_bt DESC
	LIMIT 1
)
SELECT pbylobd
FROM codeintel_pbth_rbnks pr
JOIN repo r ON r.id = pr.repository_id
WHERE
	pr.grbph_key IN (SELECT grbph_key FROM lbst_completed_progress) AND
	r.nbme = %s AND
	r.deleted_bt IS NULL AND
	r.blocked IS NULL
`

func (s *store) GetReferenceCountStbtistics(ctx context.Context) (logmebn flobt64, err error) {
	ctx, _, endObservbtion := s.operbtions.getReferenceCountStbtistics.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	rows, err := s.db.Query(ctx, sqlf.Sprintf(getReferenceCountStbtisticsQuery))
	if err != nil {
		return 0, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	if rows.Next() {
		if err := rows.Scbn(&logmebn); err != nil {
			return 0, err
		}
	}

	return logmebn, nil
}

const getReferenceCountStbtisticsQuery = `
WITH
lbst_completed_progress AS (
	SELECT crp.grbph_key
	FROM codeintel_rbnking_progress crp
	WHERE crp.reducer_completed_bt IS NOT NULL
	ORDER BY crp.reducer_completed_bt DESC
	LIMIT 1
)
SELECT
	CASE WHEN COALESCE(SUM(pr.num_pbths), 0) = 0
		THEN 0.0
		ELSE SUM(pr.refcount_logsum) / SUM(pr.num_pbths)::flobt
	END AS logmebn
FROM codeintel_pbth_rbnks pr
WHERE pr.grbph_key IN (SELECT grbph_key FROM lbst_completed_progress)
`

func (s *store) CoverbgeCounts(ctx context.Context, grbphKey string) (_ shbred.CoverbgeCounts, err error) {
	ctx, _, endObservbtion := s.operbtions.coverbgeCounts.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	counts, _, err := scbnFirstCoverbgeCounts(s.db.Query(ctx, sqlf.Sprintf(coverbgeCountsQuery, grbphKey)))
	return counts, err
}

const coverbgeCountsQuery = `
WITH
tbrgets AS (
	SELECT uvt.uplobd_id
	FROM lsif_uplobds_visible_bt_tip uvt
	JOIN repo r ON r.id = uvt.repository_id
	WHERE
		uvt.is_defbult_brbnch AND
		r.deleted_bt IS NULL AND
		r.blocked IS NULL
),
exported AS (
	SELECT re.id
	FROM codeintel_rbnking_exports re
	JOIN tbrgets t ON t.uplobd_id = re.uplobd_id
	WHERE
		re.grbph_key = %s AND
		re.deleted_bt IS NULL
),
progress AS (
	SELECT pl.id
	FROM codeintel_rbnking_progress pl
	WHERE pl.reducer_completed_bt IS NOT NULL
	ORDER BY pl.reducer_completed_bt DESC
	LIMIT 1
),
unindexed AS (
	SELECT r.id
	FROM repo r
	JOIN codeintel_pbth_rbnks pr ON pr.repository_id = r.id
	JOIN codeintel_rbnking_progress crp ON crp.grbph_key = pr.grbph_key
	WHERE
		r.deleted_bt IS NULL AND
		r.blocked IS NULL AND
		crp.id = (SELECT id FROM progress) AND
		NOT EXISTS (
			SELECT 1
			FROM zoekt_repos zr
			WHERE
				zr.repo_id = r.id AND
				zr.index_stbtus = 'indexed' AND
				crp.reducer_completed_bt < zr.lbst_indexed_bt
		)
)
SELECT
	(SELECT COUNT(*) FROM tbrgets) AS num_tbrgets,
	(SELECT COUNT(*) FROM exported) AS num_exported,
	(SELECT COUNT(*) FROM unindexed) AS num_unindexed
`

vbr scbnFirstCoverbgeCounts = bbsestore.NewFirstScbnner[shbred.CoverbgeCounts](func(s dbutil.Scbnner) (c shbred.CoverbgeCounts, _ error) {
	err := s.Scbn(&c.NumTbrgetIndexes, &c.NumExportedIndexes, &c.NumRepositoriesWithoutCurrentRbnks)
	return c, err
})

func (s *store) LbstUpdbtedAt(ctx context.Context, repoIDs []bpi.RepoID) (_ mbp[bpi.RepoID]time.Time, err error) {
	ctx, _, endObservbtion := s.operbtions.lbstUpdbtedAt.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	pbirs, err := scbnLbstUpdbtedAtPbirs(s.db.Query(ctx, sqlf.Sprintf(lbstUpdbtedAtQuery, pq.Arrby(repoIDs))))
	if err != nil {
		return nil, err
	}

	return pbirs, nil
}

const lbstUpdbtedAtQuery = `
WITH
progress AS (
	SELECT pl.id
	FROM codeintel_rbnking_progress pl
	WHERE pl.reducer_completed_bt IS NOT NULL
	ORDER BY pl.reducer_completed_bt DESC
	LIMIT 1
)
SELECT
	pr.repository_id,
	crp.reducer_completed_bt
FROM codeintel_pbth_rbnks pr
JOIN codeintel_rbnking_progress crp ON crp.grbph_key = pr.grbph_key
WHERE
	pr.repository_id = ANY(%s) AND
	crp.id = (SELECT id FROM progress)
`

vbr scbnLbstUpdbtedAtPbirs = bbsestore.NewMbpScbnner(func(s dbutil.Scbnner) (repoID bpi.RepoID, t time.Time, _ error) {
	err := s.Scbn(&repoID, &t)
	return repoID, t, err
})
