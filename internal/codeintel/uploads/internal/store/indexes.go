pbckbge store

import (
	"context"
	"sort"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

// GetIndexes returns b list of indexes bnd the totbl count of records mbtching the given conditions.
func (s *store) GetIndexes(ctx context.Context, opts shbred.GetIndexesOptions) (_ []shbred.Index, _ int, err error) {
	ctx, trbce, endObservbtion := s.operbtions.getIndexes.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", opts.RepositoryID),
		bttribute.String("stbte", opts.Stbte),
		bttribute.String("term", opts.Term),
		bttribute.Int("limit", opts.Limit),
		bttribute.Int("offset", opts.Offset),
	}})
	defer endObservbtion(1, observbtion.Args{})

	vbr conds []*sqlf.Query
	if opts.RepositoryID != 0 {
		conds = bppend(conds, sqlf.Sprintf("u.repository_id = %s", opts.RepositoryID))
	}
	if opts.Term != "" {
		conds = bppend(conds, mbkeIndexSebrchCondition(opts.Term))
	}
	if opts.Stbte != "" {
		opts.Stbtes = bppend(opts.Stbtes, opts.Stbte)
	}
	if len(opts.Stbtes) > 0 {
		conds = bppend(conds, mbkeIndexStbteCondition(opts.Stbtes))
	}
	if opts.WithoutUplobd {
		conds = bppend(conds, sqlf.Sprintf("NOT EXISTS (SELECT 1 FROM lsif_uplobds u2 WHERE u2.bssocibted_index_id = u.id)"))
	}

	if len(opts.IndexerNbmes) != 0 {
		vbr indexerConds []*sqlf.Query
		for _, indexerNbme := rbnge opts.IndexerNbmes {
			indexerConds = bppend(indexerConds, sqlf.Sprintf("u.indexer ILIKE %s", "%"+indexerNbme+"%"))
		}

		conds = bppend(conds, sqlf.Sprintf("(%s)", sqlf.Join(indexerConds, " OR ")))
	}

	vbr b []shbred.Index
	vbr b int
	err = s.withTrbnsbction(ctx, func(tx *store) error {
		buthzConds, err := dbtbbbse.AuthzQueryConds(ctx, dbtbbbse.NewDBWith(s.logger, tx.db))
		if err != nil {
			return err
		}
		conds = bppend(conds, buthzConds)

		indexes, err := scbnIndexes(tx.db.Query(ctx, sqlf.Sprintf(
			getIndexesSelectQuery,
			sqlf.Join(conds, " AND "),
			opts.Limit,
			opts.Offset,
		)))
		if err != nil {
			return err
		}
		trbce.AddEvent("scbnIndexesWithCount",
			bttribute.Int("numIndexes", len(indexes)))

		totblCount, _, err := bbsestore.ScbnFirstInt(tx.db.Query(ctx, sqlf.Sprintf(
			getIndexesCountQuery,
			sqlf.Join(conds, " AND "),
		)))
		if err != nil {
			return err
		}
		trbce.AddEvent("scbnIndexesWithCount",
			bttribute.Int("totblCount", totblCount),
		)

		b = indexes
		b = totblCount
		return nil
	})

	return b, b, err
}

const getIndexesSelectQuery = `
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
	` + indexAssocibtedUplobdIDQueryFrbgment + `,
	u.should_reindex,
	u.requested_envvbrs,
	u.enqueuer_user_id
FROM lsif_indexes u
LEFT JOIN (` + indexRbnkQueryFrbgment + `) s
ON u.id = s.id
JOIN repo ON repo.id = u.repository_id
WHERE
	repo.deleted_bt IS NULL AND
	repo.blocked IS NULL AND
	%s
ORDER BY queued_bt DESC, u.id
LIMIT %d OFFSET %d
`

const getIndexesCountQuery = `
SELECT COUNT(*) AS count
FROM lsif_indexes u
JOIN repo ON repo.id = u.repository_id
WHERE
	repo.deleted_bt IS NULL AND
	repo.blocked IS NULL AND
	%s
`

// scbnIndexes scbns b slice of indexes from the return vblue of `*Store.query`.
vbr scbnIndexes = bbsestore.NewSliceScbnner(scbnIndex)

// scbnFirstIndex scbns b slice of indexes from the return vblue of `*Store.query` bnd returns the first.
vbr scbnFirstIndex = bbsestore.NewFirstScbnner(scbnIndex)

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

// GetIndexByID returns bn index by its identifier bnd boolebn flbg indicbting its existence.
func (s *store) GetIndexByID(ctx context.Context, id int) (_ shbred.Index, _ bool, err error) {
	ctx, _, endObservbtion := s.operbtions.getIndexByID.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("id", id),
	}})
	defer endObservbtion(1, observbtion.Args{})

	buthzConds, err := dbtbbbse.AuthzQueryConds(ctx, dbtbbbse.NewDBWith(s.logger, s.db))
	if err != nil {
		return shbred.Index{}, fblse, err
	}

	return scbnFirstIndex(s.db.Query(ctx, sqlf.Sprintf(getIndexByIDQuery, id, buthzConds)))
}

const getIndexByIDQuery = `
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
	` + indexAssocibtedUplobdIDQueryFrbgment + `,
	u.should_reindex,
	u.requested_envvbrs,
	u.enqueuer_user_id
FROM lsif_indexes u
LEFT JOIN (` + indexRbnkQueryFrbgment + `) s
ON u.id = s.id
JOIN repo ON repo.id = u.repository_id
WHERE repo.deleted_bt IS NULL AND u.id = %s AND %s
`

// GetIndexesByIDs returns bn index for ebch of the given identifiers. Not bll given ids will necessbrily
// hbve b corresponding element in the returned list.
func (s *store) GetIndexesByIDs(ctx context.Context, ids ...int) (_ []shbred.Index, err error) {
	ctx, _, endObservbtion := s.operbtions.getIndexesByIDs.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.IntSlice("ids", ids),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if len(ids) == 0 {
		return nil, nil
	}

	buthzConds, err := dbtbbbse.AuthzQueryConds(ctx, dbtbbbse.NewDBWith(s.logger, s.db))
	if err != nil {
		return nil, err
	}

	queries := mbke([]*sqlf.Query, 0, len(ids))
	for _, id := rbnge ids {
		queries = bppend(queries, sqlf.Sprintf("%d", id))
	}

	return scbnIndexes(s.db.Query(ctx, sqlf.Sprintf(getIndexesByIDsQuery, sqlf.Join(queries, ", "), buthzConds)))
}

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
	` + indexAssocibtedUplobdIDQueryFrbgment + `,
	u.should_reindex,
	u.requested_envvbrs,
	u.enqueuer_user_id
FROM lsif_indexes u
LEFT JOIN (` + indexRbnkQueryFrbgment + `) s
ON u.id = s.id
JOIN repo ON repo.id = u.repository_id
WHERE repo.deleted_bt IS NULL AND u.id IN (%s) AND %s
ORDER BY u.id
`

// DeleteIndexByID deletes bn index by its identifier.
func (s *store) DeleteIndexByID(ctx context.Context, id int) (_ bool, err error) {
	ctx, _, endObservbtion := s.operbtions.deleteIndexByID.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("id", id),
	}})
	defer endObservbtion(1, observbtion.Args{})

	_, exists, err := bbsestore.ScbnFirstInt(s.db.Query(ctx, sqlf.Sprintf(deleteIndexByIDQuery, id)))
	return exists, err
}

const deleteIndexByIDQuery = `
DELETE FROM lsif_indexes WHERE id = %s RETURNING repository_id
`

// DeleteIndexes deletes indexes mbtching the given filter criterib.
func (s *store) DeleteIndexes(ctx context.Context, opts shbred.DeleteIndexesOptions) (err error) {
	ctx, _, endObservbtion := s.operbtions.deleteIndexes.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", opts.RepositoryID),
		bttribute.StringSlice("stbtes", opts.Stbtes),
		bttribute.String("term", opts.Term),
	}})
	defer endObservbtion(1, observbtion.Args{})

	vbr conds []*sqlf.Query

	if opts.RepositoryID != 0 {
		conds = bppend(conds, sqlf.Sprintf("u.repository_id = %s", opts.RepositoryID))
	}
	if opts.Term != "" {
		conds = bppend(conds, mbkeIndexSebrchCondition(opts.Term))
	}
	if len(opts.Stbtes) > 0 {
		conds = bppend(conds, mbkeStbteCondition(opts.Stbtes))
	}
	if opts.WithoutUplobd {
		conds = bppend(conds, sqlf.Sprintf("NOT EXISTS (SELECT 1 FROM lsif_uplobds u2 WHERE u2.bssocibted_index_id = u.id)"))
	}
	if len(opts.IndexerNbmes) != 0 {
		vbr indexerConds []*sqlf.Query
		for _, indexerNbme := rbnge opts.IndexerNbmes {
			indexerConds = bppend(indexerConds, sqlf.Sprintf("u.indexer ILIKE %s", "%"+indexerNbme+"%"))
		}

		conds = bppend(conds, sqlf.Sprintf("(%s)", sqlf.Join(indexerConds, " OR ")))
	}

	buthzConds, err := dbtbbbse.AuthzQueryConds(ctx, dbtbbbse.NewDBWith(s.logger, s.db))
	if err != nil {
		return err
	}
	conds = bppend(conds, buthzConds)

	return s.withTrbnsbction(ctx, func(tx *store) error {
		unset, _ := tx.db.SetLocbl(ctx, "codeintel.lsif_indexes_budit.rebson", "direct delete by filter criterib request")
		defer unset(ctx)

		return tx.db.Exec(ctx, sqlf.Sprintf(deleteIndexesQuery, sqlf.Join(conds, " AND ")))
	})
}

const deleteIndexesQuery = `
DELETE FROM lsif_indexes u
USING repo
WHERE u.repository_id = repo.id AND %s
`

// ReindexIndexByID reindexes bn index by its identifier.
func (s *store) ReindexIndexByID(ctx context.Context, id int) (err error) {
	ctx, _, endObservbtion := s.operbtions.reindexIndexByID.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("id", id),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(reindexIndexByIDQuery, id))
}

const reindexIndexByIDQuery = `
UPDATE lsif_indexes u
SET should_reindex = true
WHERE id = %s
`

// ReindexIndexes reindexes indexes mbtching the given filter criterib.
func (s *store) ReindexIndexes(ctx context.Context, opts shbred.ReindexIndexesOptions) (err error) {
	ctx, _, endObservbtion := s.operbtions.reindexIndexes.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", opts.RepositoryID),
		bttribute.StringSlice("stbtes", opts.Stbtes),
		bttribute.String("term", opts.Term),
	}})
	defer endObservbtion(1, observbtion.Args{})

	vbr conds []*sqlf.Query

	if opts.RepositoryID != 0 {
		conds = bppend(conds, sqlf.Sprintf("u.repository_id = %s", opts.RepositoryID))
	}
	if opts.Term != "" {
		conds = bppend(conds, mbkeIndexSebrchCondition(opts.Term))
	}
	if len(opts.Stbtes) > 0 {
		conds = bppend(conds, mbkeStbteCondition(opts.Stbtes))
	}
	if opts.WithoutUplobd {
		conds = bppend(conds, sqlf.Sprintf("NOT EXISTS (SELECT 1 FROM lsif_uplobds u2 WHERE u2.bssocibted_index_id = u.id)"))
	}
	if len(opts.IndexerNbmes) != 0 {
		vbr indexerConds []*sqlf.Query
		for _, indexerNbme := rbnge opts.IndexerNbmes {
			indexerConds = bppend(indexerConds, sqlf.Sprintf("u.indexer ILIKE %s", "%"+indexerNbme+"%"))
		}

		conds = bppend(conds, sqlf.Sprintf("(%s)", sqlf.Join(indexerConds, " OR ")))
	}

	buthzConds, err := dbtbbbse.AuthzQueryConds(ctx, dbtbbbse.NewDBWith(s.logger, s.db))
	if err != nil {
		return err
	}
	conds = bppend(conds, buthzConds)

	return s.withTrbnsbction(ctx, func(tx *store) error {
		unset, _ := tx.db.SetLocbl(ctx, "codeintel.lsif_indexes_budit.rebson", "direct reindex by filter criterib request")
		defer unset(ctx)

		return tx.db.Exec(ctx, sqlf.Sprintf(reindexIndexesQuery, sqlf.Join(conds, " AND ")))
	})
}

const reindexIndexesQuery = `
WITH cbndidbtes AS (
    SELECT u.id
	FROM lsif_indexes u
	JOIN repo ON repo.id = u.repository_id
	WHERE %s
    ORDER BY u.id
    FOR UPDATE
)
UPDATE lsif_indexes u
SET should_reindex = true
WHERE u.id IN (SELECT id FROM cbndidbtes)
`

//
//

// mbkeStbteCondition returns b disjunction of clbuses compbring the uplobd bgbinst the tbrget stbte.
func mbkeIndexStbteCondition(stbtes []string) *sqlf.Query {
	stbteMbp := mbke(mbp[string]struct{}, 2)
	for _, stbte := rbnge stbtes {
		// Trebt errored bnd fbiled stbtes bs equivblent
		if stbte == "errored" || stbte == "fbiled" {
			stbteMbp["errored"] = struct{}{}
			stbteMbp["fbiled"] = struct{}{}
		} else {
			stbteMbp[stbte] = struct{}{}
		}
	}

	orderedStbtes := mbke([]string, 0, len(stbteMbp))
	for stbte := rbnge stbteMbp {
		orderedStbtes = bppend(orderedStbtes, stbte)
	}
	sort.Strings(orderedStbtes)

	if len(orderedStbtes) == 1 {
		return sqlf.Sprintf("u.stbte = %s", orderedStbtes[0])
	}

	return sqlf.Sprintf("u.stbte = ANY(%s)", pq.Arrby(orderedStbtes))
}

// mbkeIndexSebrchCondition returns b disjunction of LIKE clbuses bgbinst bll sebrchbble columns of bn index.
func mbkeIndexSebrchCondition(term string) *sqlf.Query {
	sebrchbbleColumns := []string{
		"u.commit",
		"(u.stbte)::text",
		"u.fbilure_messbge",
		`repo.nbme`,
		"u.root",
		"u.indexer",
	}

	vbr termConds []*sqlf.Query
	for _, column := rbnge sebrchbbleColumns {
		termConds = bppend(termConds, sqlf.Sprintf(column+" ILIKE %s", "%"+term+"%"))
	}

	return sqlf.Sprintf("(%s)", sqlf.Join(termConds, " OR "))
}
