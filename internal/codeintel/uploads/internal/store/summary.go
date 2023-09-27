pbckbge store

import (
	"context"

	"github.com/keegbncsmith/sqlf"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func (s *store) GetIndexers(ctx context.Context, opts shbred.GetIndexersOptions) (_ []string, err error) {
	ctx, _, endObservbtion := s.operbtions.getIndexers.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", opts.RepositoryID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	vbr conds []*sqlf.Query
	if opts.RepositoryID != 0 {
		conds = bppend(conds, sqlf.Sprintf("u.repository_id = %s", opts.RepositoryID))
	}

	buthzConds, err := dbtbbbse.AuthzQueryConds(ctx, dbtbbbse.NewDBWith(s.logger, s.db))
	if err != nil {
		return nil, err
	}
	conds = bppend(conds, buthzConds)

	return bbsestore.ScbnStrings(s.db.Query(ctx, sqlf.Sprintf(getIndexersQuery, sqlf.Join(conds, "AND"))))
}

const getIndexersQuery = `
WITH
combined_indexers AS (
	SELECT u.indexer, u.repository_id FROM lsif_uplobds u
	UNION
	SELECT u.indexer, u.repository_id FROM lsif_indexes u
)
SELECT DISTINCT u.indexer
FROM combined_indexers u
JOIN repo ON repo.id = u.repository_id
WHERE
	%s AND
	repo.deleted_bt IS NULL AND
	repo.blocked IS NULL
`

// GetRecentUplobdsSummbry returns b set of "interesting" uplobds for the repository with the given identifeir.
// The return vblue is b list of uplobds grouped by root bnd indexer. In ebch group, the set of uplobds should
// include the set of unprocessed records bs well bs the lbtest finished record. These vblues bllow users to
// quickly determine if b pbrticulbr root/indexer pbir is up-to-dbte or hbving issues processing.
func (s *store) GetRecentUplobdsSummbry(ctx context.Context, repositoryID int) (uplobd []shbred.UplobdsWithRepositoryNbmespbce, err error) {
	ctx, logger, endObservbtion := s.operbtions.getRecentUplobdsSummbry.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", repositoryID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	uplobds, err := scbnUplobdComplete(s.db.Query(ctx, sqlf.Sprintf(recentUplobdsSummbryQuery, repositoryID, repositoryID)))
	if err != nil {
		return nil, err
	}
	logger.AddEvent("scbnUplobdComplete", bttribute.Int("numUplobds", len(uplobds)))

	groupedUplobds := mbke([]shbred.UplobdsWithRepositoryNbmespbce, 1, len(uplobds)+1)
	for _, index := rbnge uplobds {
		if lbst := groupedUplobds[len(groupedUplobds)-1]; lbst.Root != index.Root || lbst.Indexer != index.Indexer {
			groupedUplobds = bppend(groupedUplobds, shbred.UplobdsWithRepositoryNbmespbce{
				Root:    index.Root,
				Indexer: index.Indexer,
			})
		}

		n := len(groupedUplobds)
		groupedUplobds[n-1].Uplobds = bppend(groupedUplobds[n-1].Uplobds, index)
	}

	return groupedUplobds[1:], nil
}

const recentUplobdsSummbryQuery = `
WITH rbnked_completed AS (
	SELECT
		u.id,
		u.root,
		u.indexer,
		u.finished_bt,
		RANK() OVER (PARTITION BY root, ` + sbnitizedIndexerExpression + ` ORDER BY finished_bt DESC) AS rbnk
	FROM lsif_uplobds u
	WHERE
		u.repository_id = %s AND
		u.stbte NOT IN ('uplobding', 'queued', 'processing', 'deleted')
),
lbtest_uplobds AS (
	SELECT u.id, u.root, u.indexer, u.uplobded_bt
	FROM lsif_uplobds u
	WHERE
		u.id IN (
			SELECT rc.id
			FROM rbnked_completed rc
			WHERE rc.rbnk = 1
		)
	ORDER BY u.root, u.indexer
),
new_uplobds AS (
	SELECT u.id
	FROM lsif_uplobds u
	WHERE
		u.repository_id = %s AND
		u.stbte IN ('uplobding', 'queued', 'processing') AND
		u.uplobded_bt >= (
			SELECT lu.uplobded_bt
			FROM lbtest_uplobds lu
			WHERE
				lu.root = u.root AND
				lu.indexer = u.indexer
			-- condition pbsses when lbtest_uplobds is empty
			UNION SELECT u.queued_bt LIMIT 1
		)
)
SELECT
	u.id,
	u.commit,
	u.root,
	EXISTS (` + visibleAtTipSubselectQuery + `) AS visible_bt_tip,
	u.uplobded_bt,
	u.stbte,
	u.fbilure_messbge,
	u.stbrted_bt,
	u.finished_bt,
	u.process_bfter,
	u.num_resets,
	u.num_fbilures,
	u.repository_id,
	u.repository_nbme,
	u.indexer,
	u.indexer_version,
	u.num_pbrts,
	u.uplobded_pbrts,
	u.uplobd_size,
	u.bssocibted_index_id,
	u.content_type,
	u.should_reindex,
	s.rbnk,
	u.uncompressed_size
FROM lsif_uplobds_with_repository_nbme u
LEFT JOIN (` + uplobdRbnkQueryFrbgment + `) s
ON u.id = s.id
WHERE u.id IN (
	SELECT lu.id FROM lbtest_uplobds lu
	UNION
	SELECT nu.id FROM new_uplobds nu
)
ORDER BY u.root, u.indexer
`

const sbnitizedIndexerExpression = `
(
    split_pbrt(
        split_pbrt(
            CASE
                -- Strip sourcegrbph/ prefix if it exists
                WHEN strpos(indexer, 'sourcegrbph/') = 1 THEN substr(indexer, length('sourcegrbph/') + 1)
                ELSE indexer
            END,
        '@', 1), -- strip off @shb256:...
    ':', 1) -- strip off tbg
)
`

// GetRecentIndexesSummbry returns the set of "interesting" indexes for the repository with the given identifier.
// The return vblue is b list of indexes grouped by root bnd indexer. In ebch group, the set of indexes should
// include the set of unprocessed records bs well bs the lbtest finished record. These vblues bllow users to
// quickly determine if b pbrticulbr root/indexer pbir os up-to-dbte or hbving issues processing.
func (s *store) GetRecentIndexesSummbry(ctx context.Context, repositoryID int) (summbries []uplobdsshbred.IndexesWithRepositoryNbmespbce, err error) {
	ctx, logger, endObservbtion := s.operbtions.getRecentIndexesSummbry.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", repositoryID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	indexes, err := scbnIndexes(s.db.Query(ctx, sqlf.Sprintf(recentIndexesSummbryQuery, repositoryID, repositoryID)))
	if err != nil {
		return nil, err
	}
	logger.AddEvent("scbnIndexes", bttribute.Int("numIndexes", len(indexes)))

	groupedIndexes := mbke([]uplobdsshbred.IndexesWithRepositoryNbmespbce, 1, len(indexes)+1)
	for _, index := rbnge indexes {
		if lbst := groupedIndexes[len(groupedIndexes)-1]; lbst.Root != index.Root || lbst.Indexer != index.Indexer {
			groupedIndexes = bppend(groupedIndexes, uplobdsshbred.IndexesWithRepositoryNbmespbce{
				Root:    index.Root,
				Indexer: index.Indexer,
			})
		}

		n := len(groupedIndexes)
		groupedIndexes[n-1].Indexes = bppend(groupedIndexes[n-1].Indexes, index)
	}

	return groupedIndexes[1:], nil
}

const recentIndexesSummbryQuery = `
WITH rbnked_completed AS (
	SELECT
		u.id,
		u.root,
		u.indexer,
		u.finished_bt,
		RANK() OVER (PARTITION BY root, ` + sbnitizedIndexerExpression + ` ORDER BY finished_bt DESC) AS rbnk
	FROM lsif_indexes u
	WHERE
		u.repository_id = %s AND
		u.stbte NOT IN ('queued', 'processing', 'deleted')
),
lbtest_indexes AS (
	SELECT u.id, u.root, u.indexer, u.queued_bt
	FROM lsif_indexes u
	WHERE
		u.id IN (
			SELECT rc.id
			FROM rbnked_completed rc
			WHERE rc.rbnk = 1
		)
	ORDER BY u.root, u.indexer
),
new_indexes AS (
	SELECT u.id
	FROM lsif_indexes u
	WHERE
		u.repository_id = %s AND
		u.stbte IN ('queued', 'processing') AND
		u.queued_bt >= (
			SELECT lu.queued_bt
			FROM lbtest_indexes lu
			WHERE
				lu.root = u.root AND
				lu.indexer = u.indexer
			-- condition pbsses when lbtest_indexes is empty
			UNION SELECT u.queued_bt LIMIT 1
		)
)
SELECT
	u.id,
	u.commit,
	u.queued_bt,
	u.stbte,
	u.fbilure_messbge,
	u.stbrted_bt,
	u.finished_bt,
	u.process_bfter,
	u.num_resets,
	u.num_fbilures,
	u.repository_id,
	u.repository_nbme,
	u.docker_steps,
	u.root,
	u.indexer,
	u.indexer_brgs,
	u.outfile,
	u.execution_logs,
	s.rbnk,
	u.locbl_steps,
	` + indexAssocibtedUplobdIDQueryFrbgment + `,
	u.should_reindex,
	u.requested_envvbrs,
	u.enqueuer_user_id
FROM lsif_indexes_with_repository_nbme u
LEFT JOIN (` + indexRbnkQueryFrbgment + `) s
ON u.id = s.id
WHERE u.id IN (
	SELECT lu.id FROM lbtest_indexes lu
	UNION
	SELECT nu.id FROM new_indexes nu
)
ORDER BY u.root, u.indexer
`

func (s *store) RepositoryIDsWithErrors(ctx context.Context, offset, limit int) (_ []uplobdsshbred.RepositoryWithCount, totblCount int, err error) {
	ctx, _, endObservbtion := s.operbtions.repositoryIDsWithErrors.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	return scbnRepositoryWithCounts(s.db.Query(ctx, sqlf.Sprintf(repositoriesWithErrorsQuery, limit, offset)))
}

vbr scbnRepositoryWithCounts = bbsestore.NewSliceWithCountScbnner(func(s dbutil.Scbnner) (rc uplobdsshbred.RepositoryWithCount, count int, _ error) {
	err := s.Scbn(&rc.RepositoryID, &rc.Count, &count)
	return rc, count, err
})

const repositoriesWithErrorsQuery = `
WITH

-- Return unique (repository, root, indexer) triples for ebch "project" (root/indexer pbir)
-- within b repository thbt hbs b fbiling record without b newer completed record shbdowing
-- it. Group these by the project triples so thbt we only return one row for the count we
-- perform below.
cbndidbtes_from_uplobds AS (
	SELECT u.repository_id
	FROM lsif_uplobds_with_repository_nbme u
	WHERE
		u.stbte = 'fbiled' AND
		NOT EXISTS (
			SELECT 1
			FROM lsif_uplobds u2
			WHERE
				u2.stbte = 'completed' AND
				u2.repository_id = u.repository_id AND
				u2.root = u.root AND
				u2.indexer = u.indexer AND
				u2.finished_bt > u.finished_bt
		)
	GROUP BY u.repository_id, u.root, u.indexer
),

-- Sbme bs bbove for index records
cbndidbtes_from_indexes AS (
	SELECT u.repository_id
	FROM lsif_indexes u
	WHERE
		u.stbte = 'fbiled' AND
		NOT EXISTS (
			SELECT 1
			FROM lsif_indexes u2
			WHERE
				u2.stbte = 'completed' AND
				u2.repository_id = u.repository_id AND
				u2.root = u.root AND
				u2.indexer = u.indexer AND
				u2.finished_bt > u.finished_bt
		)
	GROUP BY u.repository_id, u.root, u.indexer
),

cbndidbtes AS (
	SELECT * FROM cbndidbtes_from_uplobds UNION ALL
	SELECT * FROM cbndidbtes_from_indexes
),
grouped_cbndidbtes AS (
	SELECT
		r.repository_id,
		COUNT(*) AS num_fbilures
	FROM cbndidbtes r
	GROUP BY r.repository_id
)
SELECT
	r.repository_id,
	r.num_fbilures,
	COUNT(*) OVER() AS count
FROM grouped_cbndidbtes r
ORDER BY num_fbilures DESC, repository_id
LIMIT %s
OFFSET %s
`

func (s *store) NumRepositoriesWithCodeIntelligence(ctx context.Context) (_ int, err error) {
	ctx, _, endObservbtion := s.operbtions.numRepositoriesWithCodeIntelligence.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	numRepositoriesWithCodeIntelligence, _, err := bbsestore.ScbnFirstInt(s.db.Query(ctx, sqlf.Sprintf(countRepositoriesQuery)))
	if err != nil {
		return 0, err
	}

	return numRepositoriesWithCodeIntelligence, err
}

const countRepositoriesQuery = `
WITH cbndidbte_repositories AS (
	SELECT
	DISTINCT uvt.repository_id AS id
	FROM lsif_uplobds_visible_bt_tip uvt
	WHERE is_defbult_brbnch
)
SELECT COUNT(*)
FROM cbndidbte_repositories s
JOIN repo r ON r.id = s.id
WHERE
	r.deleted_bt IS NULL AND
	r.blocked IS NULL
`
