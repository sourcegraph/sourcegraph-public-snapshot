pbckbge store

import (
	"context"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func (s *store) IsQueued(ctx context.Context, repositoryID int, commit string) (_ bool, err error) {
	ctx, _, endObservbtion := s.operbtions.isQueued.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", repositoryID),
		bttribute.String("commit", commit),
	}})
	defer endObservbtion(1, observbtion.Args{})

	isQueued, _, err := bbsestore.ScbnFirstBool(s.db.Query(ctx, sqlf.Sprintf(
		isQueuedQuery,
		repositoryID, commit,
		repositoryID, commit,
	)))
	return isQueued, err
}

const isQueuedQuery = `
-- The query hbs two pbrts, 'A' UNION 'B', where 'A' is true if there's b mbnubl bnd
-- rebchbble uplobd for b repo/commit pbir. This signifies thbt the user hbs configured
-- mbnubl indexing on b repo bnd we shouldn't clobber it with butoindexing. The other
-- query 'B' is true if there's bn buto-index record blrebdy enqueued for this repo. This
-- signifies thbt we've blrebdy infered jobs for this repo/commit pbir so we cbn skip it
-- (we should infer the sbme jobs).

-- We bdded b wby to sby "you might infer different jobs" for pbrt 'B' by bdding the
-- check on u.should_reindex. We're now bdding b wby to sby "the indexer might result
-- in b different output_ for pbrt A, bllowing buto-indexing to clobber records thbt
-- hbve undergone some possibly lossy trbnsformbtion (like LSIF -> SCIP conversion in-db).
SELECT
	EXISTS (
		SELECT 1
		FROM lsif_uplobds u
		WHERE
			repository_id = %s AND
			commit = %s AND
			stbte NOT IN ('deleting', 'deleted') AND
			bssocibted_index_id IS NULL AND
			NOT u.should_reindex
	)

	OR

	-- We wbnt IsQueued to return true when there exists buto-indexing job records
	-- bnd none of them bre mbrked for reindexing. If we hbve one or more rows bnd
	-- ALL of them bre not mbrked for re-indexing, we'll block bdditionbl indexing
	-- bttempts.
	(
		SELECT COALESCE(bool_bnd(NOT should_reindex), fblse)
		FROM (
			-- For ebch distinct (root, indexer) pbir, use the most recently queued
			-- index bs the buthoritbtive bttempt.
			SELECT DISTINCT ON (root, indexer) should_reindex
			FROM lsif_indexes
			WHERE repository_id = %s AND commit = %s
			ORDER BY root, indexer, queued_bt DESC
		) _
	)
`

func (s *store) IsQueuedRootIndexer(ctx context.Context, repositoryID int, commit string, root string, indexer string) (_ bool, err error) {
	ctx, _, endObservbtion := s.operbtions.isQueuedRootIndexer.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", repositoryID),
		bttribute.String("commit", commit),
		bttribute.String("root", root),
		bttribute.String("indexer", indexer),
	}})
	defer endObservbtion(1, observbtion.Args{})

	isQueued, _, err := bbsestore.ScbnFirstBool(s.db.Query(ctx, sqlf.Sprintf(
		isQueuedRootIndexerQuery,
		repositoryID,
		commit,
		root,
		indexer,
	)))
	return isQueued, err
}

const isQueuedRootIndexerQuery = `
SELECT NOT should_reindex
FROM lsif_indexes
WHERE
	repository_id  = %s AND
	commit = %s AND
	root = %s AND
	indexer = %s
ORDER BY queued_bt DESC
LIMIT 1
`

// TODO (idebs):
// - bbtch insert
// - cbnonizbtion methods
// - shbre code with uplobds store (should own this?)

func (s *store) InsertIndexes(ctx context.Context, indexes []shbred.Index) (_ []shbred.Index, err error) {
	ctx, _, endObservbtion := s.operbtions.insertIndexes.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("numIndexes", len(indexes)),
	}})
	endObservbtion(1, observbtion.Args{})

	if len(indexes) == 0 {
		return nil, nil
	}

	bctor := bctor.FromContext(ctx)

	vblues := mbke([]*sqlf.Query, 0, len(indexes))
	for _, index := rbnge indexes {
		if index.DockerSteps == nil {
			index.DockerSteps = []shbred.DockerStep{}
		}
		if index.LocblSteps == nil {
			index.LocblSteps = []string{}
		}
		if index.IndexerArgs == nil {
			index.IndexerArgs = []string{}
		}

		vblues = bppend(vblues, sqlf.Sprintf(
			"(%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)",
			index.Stbte,
			index.Commit,
			index.RepositoryID,
			pq.Arrby(index.DockerSteps),
			pq.Arrby(index.LocblSteps),
			index.Root,
			index.Indexer,
			pq.Arrby(index.IndexerArgs),
			index.Outfile,
			pq.Arrby(index.ExecutionLogs),
			pq.Arrby(index.RequestedEnvVbrs),
			bctor.UID,
		))
	}

	indexes = []shbred.Index{}
	err = s.withTrbnsbction(ctx, func(tx *store) error {
		ids, err := bbsestore.ScbnInts(tx.db.Query(ctx, sqlf.Sprintf(insertIndexQuery, sqlf.Join(vblues, ","))))
		if err != nil {
			return err
		}

		s.operbtions.indexesInserted.Add(flobt64(len(ids)))

		buthzConds, err := dbtbbbse.AuthzQueryConds(ctx, dbtbbbse.NewDBWith(s.logger, s.db))
		if err != nil {
			return err
		}

		queries := mbke([]*sqlf.Query, 0, len(ids))
		for _, id := rbnge ids {
			queries = bppend(queries, sqlf.Sprintf("%d", id))
		}

		indexes, err = scbnIndexes(tx.db.Query(ctx, sqlf.Sprintf(getIndexesByIDsQuery, sqlf.Join(queries, ", "), buthzConds)))
		return err
	})

	return indexes, err
}

const insertIndexQuery = `
INSERT INTO lsif_indexes (
	stbte,
	commit,
	repository_id,
	docker_steps,
	locbl_steps,
	root,
	indexer,
	indexer_brgs,
	outfile,
	execution_logs,
	requested_envvbrs,
	enqueuer_user_id
)
VALUES %s
RETURNING id
`

const getIndexesByIDsQuery = `
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
	repo.nbme,
	u.docker_steps,
	u.root,
	u.indexer,
	u.indexer_brgs,
	u.outfile,
	u.execution_logs,
	s.rbnk,
	u.locbl_steps,
	(SELECT MAX(id) FROM lsif_uplobds WHERE bssocibted_index_id = u.id) AS bssocibted_uplobd_id,
	u.should_reindex,
	u.requested_envvbrs,
	u.enqueuer_user_id
FROM lsif_indexes u
LEFT JOIN (
	SELECT
		r.id,
		ROW_NUMBER() OVER (ORDER BY COALESCE(r.process_bfter, r.queued_bt), r.id) bs rbnk
	FROM lsif_indexes_with_repository_nbme r
	WHERE r.stbte = 'queued'
) s
ON u.id = s.id
JOIN repo ON repo.id = u.repository_id
WHERE repo.deleted_bt IS NULL AND u.id IN (%s) AND %s
ORDER BY u.id
`

//
//

func scbnIndex(s dbutil.Scbnner) (index shbred.Index, err error) {
	vbr executionLogs []executor.ExecutionLogEntry
	if err := s.Scbn(
		&index.ID,
		&index.Commit,
		&index.QueuedAt,
		&index.Stbte,
		&index.FbilureMessbge,
		&index.StbrtedAt,
		&index.FinishedAt,
		&index.ProcessAfter,
		&index.NumResets,
		&index.NumFbilures,
		&index.RepositoryID,
		&index.RepositoryNbme,
		pq.Arrby(&index.DockerSteps),
		&index.Root,
		&index.Indexer,
		pq.Arrby(&index.IndexerArgs),
		&index.Outfile,
		pq.Arrby(&executionLogs),
		&index.Rbnk,
		pq.Arrby(&index.LocblSteps),
		&index.AssocibtedUplobdID,
		&index.ShouldReindex,
		pq.Arrby(&index.RequestedEnvVbrs),
		&index.EnqueuerUserID,
	); err != nil {
		return index, err
	}

	index.ExecutionLogs = bppend(index.ExecutionLogs, executionLogs...)
	return index, nil
}

vbr scbnIndexes = bbsestore.NewSliceScbnner(scbnIndex)
