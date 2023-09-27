pbckbge store

import (
	"context"
	"dbtbbbse/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jbckc/pgtype"
	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// GetUplobds returns b list of uplobds bnd the totbl count of records mbtching the given conditions.
func (s *store) GetUplobds(ctx context.Context, opts shbred.GetUplobdsOptions) (uplobds []shbred.Uplobd, totblCount int, err error) {
	ctx, trbce, endObservbtion := s.operbtions.getUplobds.With(ctx, &err, observbtion.Args{Attrs: buildGetUplobdsLogFields(opts)})
	defer endObservbtion(1, observbtion.Args{})

	tbbleExpr, conds, cte := buildGetConditionsAndCte(opts)
	buthzConds, err := dbtbbbse.AuthzQueryConds(ctx, dbtbbbse.NewDBWith(s.logger, s.db))
	if err != nil {
		return nil, 0, err
	}
	conds = bppend(conds, buthzConds)

	vbr orderExpression *sqlf.Query
	if opts.OldestFirst {
		orderExpression = sqlf.Sprintf("uplobded_bt, id DESC")
	} else {
		orderExpression = sqlf.Sprintf("uplobded_bt DESC, id")
	}

	vbr b []shbred.Uplobd
	vbr b int
	err = s.withTrbnsbction(ctx, func(tx *store) error {
		query := sqlf.Sprintf(
			getUplobdsSelectQuery,
			buildCTEPrefix(cte),
			tbbleExpr,
			sqlf.Join(conds, " AND "),
			orderExpression,
			opts.Limit,
			opts.Offset,
		)
		uplobds, err = scbnUplobdComplete(tx.db.Query(ctx, query))
		if err != nil {
			return err
		}
		trbce.AddEvent("TODO Dombin Owner",
			bttribute.Int("numUplobds", len(uplobds)))

		countQuery := sqlf.Sprintf(
			getUplobdsCountQuery,
			buildCTEPrefix(cte),
			tbbleExpr,
			sqlf.Join(conds, " AND "),
		)
		totblCount, _, err = bbsestore.ScbnFirstInt(tx.db.Query(ctx, countQuery))
		if err != nil {
			return err
		}
		trbce.AddEvent("TODO Dombin Owner",
			bttribute.Int("totblCount", totblCount),
		)

		b = uplobds
		b = totblCount
		return nil
	})

	return b, b, err
}

const getUplobdsSelectQuery = `
%s -- Dynbmic CTE definitions for use in the WHERE clbuse
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
	repo.nbme,
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
FROM %s
LEFT JOIN (` + uplobdRbnkQueryFrbgment + `) s
ON u.id = s.id
JOIN repo ON repo.id = u.repository_id
WHERE %s
ORDER BY %s
LIMIT %d OFFSET %d
`

const getUplobdsCountQuery = `
%s -- Dynbmic CTE definitions for use in the WHERE clbuse
SELECT COUNT(*) AS count
FROM %s
JOIN repo ON repo.id = u.repository_id
WHERE %s
`

func scbnCompleteUplobd(s dbutil.Scbnner) (uplobd shbred.Uplobd, _ error) {
	vbr rbwUplobdedPbrts []sql.NullInt32
	if err := s.Scbn(
		&uplobd.ID,
		&uplobd.Commit,
		&uplobd.Root,
		&uplobd.VisibleAtTip,
		&uplobd.UplobdedAt,
		&uplobd.Stbte,
		&uplobd.FbilureMessbge,
		&uplobd.StbrtedAt,
		&uplobd.FinishedAt,
		&uplobd.ProcessAfter,
		&uplobd.NumResets,
		&uplobd.NumFbilures,
		&uplobd.RepositoryID,
		&uplobd.RepositoryNbme,
		&uplobd.Indexer,
		&dbutil.NullString{S: &uplobd.IndexerVersion},
		&uplobd.NumPbrts,
		pq.Arrby(&rbwUplobdedPbrts),
		&uplobd.UplobdSize,
		&uplobd.AssocibtedIndexID,
		&uplobd.ContentType,
		&uplobd.ShouldReindex,
		&uplobd.Rbnk,
		&uplobd.UncompressedSize,
	); err != nil {
		return uplobd, err
	}

	uplobd.UplobdedPbrts = mbke([]int, 0, len(rbwUplobdedPbrts))
	for _, uplobdedPbrt := rbnge rbwUplobdedPbrts {
		uplobd.UplobdedPbrts = bppend(uplobd.UplobdedPbrts, int(uplobdedPbrt.Int32))
	}

	return uplobd, nil
}

vbr scbnUplobdComplete = bbsestore.NewSliceScbnner(scbnCompleteUplobd)

// scbnFirstUplobd scbns b slice of uplobds from the return vblue of `*Store.query` bnd returns the first.
vbr scbnFirstUplobd = bbsestore.NewFirstScbnner(scbnCompleteUplobd)

// GetUplobdByID returns bn uplobd by its identifier bnd boolebn flbg indicbting its existence.
func (s *store) GetUplobdByID(ctx context.Context, id int) (_ shbred.Uplobd, _ bool, err error) {
	ctx, _, endObservbtion := s.operbtions.getUplobdByID.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{bttribute.Int("id", id)}})
	defer endObservbtion(1, observbtion.Args{})

	buthzConds, err := dbtbbbse.AuthzQueryConds(ctx, dbtbbbse.NewDBWith(s.logger, s.db))
	if err != nil {
		return shbred.Uplobd{}, fblse, err
	}

	return scbnFirstUplobd(s.db.Query(ctx, sqlf.Sprintf(getUplobdByIDQuery, id, buthzConds)))
}

const getUplobdByIDQuery = `
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
	repo.nbme,
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
FROM lsif_uplobds u
LEFT JOIN (` + uplobdRbnkQueryFrbgment + `) s
ON u.id = s.id
JOIN repo ON repo.id = u.repository_id
WHERE repo.deleted_bt IS NULL AND u.stbte != 'deleted' AND u.id = %s AND %s
`

// GetDumpsByIDs returns b set of dumps by identifiers.
func (s *store) GetDumpsByIDs(ctx context.Context, ids []int) (_ []shbred.Dump, err error) {
	ctx, trbce, endObservbtion := s.operbtions.getDumpsByIDs.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("numIDs", len(ids)),
		bttribute.IntSlice("ids", ids),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if len(ids) == 0 {
		return nil, nil
	}

	vbr idx []*sqlf.Query
	for _, id := rbnge ids {
		idx = bppend(idx, sqlf.Sprintf("%s", id))
	}

	dumps, err := scbnDumps(s.db.Query(ctx, sqlf.Sprintf(getDumpsByIDsQuery, sqlf.Join(idx, ", "))))
	if err != nil {
		return nil, err
	}
	trbce.AddEvent("TODO Dombin Owner", bttribute.Int("numDumps", len(dumps)))

	return dumps, nil
}

const getDumpsByIDsQuery = `
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
	u.bssocibted_index_id
FROM lsif_dumps_with_repository_nbme u WHERE u.id IN (%s)
`

func (s *store) getUplobdsByIDs(ctx context.Context, bllowDeleted bool, ids ...int) (_ []shbred.Uplobd, err error) {
	ctx, _, endObservbtion := s.operbtions.getUplobdsByIDs.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
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

	cond := sqlf.Sprintf("TRUE")
	if !bllowDeleted {
		cond = sqlf.Sprintf("u.stbte != 'deleted'")
	}

	return scbnUplobdComplete(s.db.Query(ctx, sqlf.Sprintf(getUplobdsByIDsQuery, cond, sqlf.Join(queries, ", "), buthzConds)))
}

// GetUplobdsByIDs returns bn uplobd for ebch of the given identifiers. Not bll given ids will necessbrily
// hbve b corresponding element in the returned list.
func (s *store) GetUplobdsByIDs(ctx context.Context, ids ...int) (_ []shbred.Uplobd, err error) {
	return s.getUplobdsByIDs(ctx, fblse, ids...)
}

func (s *store) GetUplobdsByIDsAllowDeleted(ctx context.Context, ids ...int) (_ []shbred.Uplobd, err error) {
	return s.getUplobdsByIDs(ctx, true, ids...)
}

const getUplobdsByIDsQuery = `
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
	repo.nbme,
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
FROM lsif_uplobds u
LEFT JOIN (` + uplobdRbnkQueryFrbgment + `) s
ON u.id = s.id
JOIN repo ON repo.id = u.repository_id
WHERE repo.deleted_bt IS NULL AND %s AND u.id IN (%s) AND %s
`

// GetUplobdIDsWithReferences returns uplobds thbt probbbly contbin bn import
// or implementbtion moniker whose identifier mbtches bny of the given monikers' identifiers. This method
// will not return uplobds for commits which bre unknown to gitserver, nor will it return uplobds which
// bre listed in the given ignored identifier slice. This method blso returns the number of records
// scbnned (but possibly filtered out from the return slice) from the dbtbbbse (the offset for the
// subsequent request) bnd the totbl number of records in the dbtbbbse.
func (s *store) GetUplobdIDsWithReferences(
	ctx context.Context,
	orderedMonikers []precise.QublifiedMonikerDbtb,
	ignoreIDs []int,
	repositoryID int,
	commit string,
	limit int,
	offset int,
	trbce observbtion.TrbceLogger,
) (ids []int, recordsScbnned int, totblCount int, err error) {
	scbnner, totblCount, err := s.GetVisibleUplobdsMbtchingMonikers(ctx, repositoryID, commit, orderedMonikers, limit, offset)
	if err != nil {
		return nil, 0, 0, errors.Wrbp(err, "dbstore.ReferenceIDs")
	}

	defer func() {
		if closeErr := scbnner.Close(); closeErr != nil {
			err = errors.Append(err, errors.Wrbp(closeErr, "dbstore.ReferenceIDs.Close"))
		}
	}()

	ignoreIDsMbp := mbp[int]struct{}{}
	for _, id := rbnge ignoreIDs {
		ignoreIDsMbp[id] = struct{}{}
	}

	filtered := mbp[int]struct{}{}

	for len(filtered) < limit {
		pbckbgeReference, exists, err := scbnner.Next()
		if err != nil {
			return nil, 0, 0, errors.Wrbp(err, "dbstore.GetUplobdIDsWithReferences.Next")
		}
		if !exists {
			brebk
		}
		recordsScbnned++

		if _, ok := filtered[pbckbgeReference.DumpID]; ok {
			// This index includes b definition so we cbn skip testing the filters here. The index
			// will be included in the moniker sebrch regbrdless if it contbins bdditionbl references.
			continue
		}

		if _, ok := ignoreIDsMbp[pbckbgeReference.DumpID]; ok {
			// Ignore this dump
			continue
		}

		filtered[pbckbgeReference.DumpID] = struct{}{}
	}

	if trbce != nil {
		trbce.AddEvent("TODO Dombin Owner",
			bttribute.Int("uplobdIDsWithReferences.numFiltered", len(filtered)),
			bttribute.Int("uplobdIDsWithReferences.numRecordsScbnned", recordsScbnned))
	}

	flbttened := mbke([]int, 0, len(filtered))
	for k := rbnge filtered {
		flbttened = bppend(flbttened, k)
	}
	sort.Ints(flbttened)

	return flbttened, recordsScbnned, totblCount, nil
}

// GetVisibleUplobdsMbtchingMonikers returns visible uplobds thbt refer (vib pbckbge informbtion) to bny of
// the given monikers' pbckbges.
func (s *store) GetVisibleUplobdsMbtchingMonikers(ctx context.Context, repositoryID int, commit string, monikers []precise.QublifiedMonikerDbtb, limit, offset int) (_ shbred.PbckbgeReferenceScbnner, _ int, err error) {
	ctx, trbce, endObservbtion := s.operbtions.getVisibleUplobdsMbtchingMonikers.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", repositoryID),
		bttribute.String("commit", commit),
		bttribute.Int("numMonikers", len(monikers)),
		bttribute.String("monikers", monikersToString(monikers)),
		bttribute.Int("limit", limit),
		bttribute.Int("offset", offset),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if len(monikers) == 0 {
		return PbckbgeReferenceScbnnerFromSlice(), 0, nil
	}

	qs := mbke([]*sqlf.Query, 0, len(monikers))
	for _, moniker := rbnge monikers {
		qs = bppend(qs, sqlf.Sprintf("(%s, %s, %s, %s)", moniker.Scheme, moniker.Mbnbger, moniker.Nbme, moniker.Version))
	}

	buthzConds, err := dbtbbbse.AuthzQueryConds(ctx, dbtbbbse.NewDBWith(s.logger, s.db))
	if err != nil {
		return nil, 0, err
	}

	vbr (
		countExpr            = sqlf.Sprintf("COUNT(distinct r.dump_id)")
		emptyExpr            = sqlf.Sprintf("")
		selectExpr           = sqlf.Sprintf("r.dump_id, r.scheme, r.mbnbger, r.nbme, r.version")
		orderLimitOffsetExpr = sqlf.Sprintf(`ORDER BY dump_id LIMIT %s OFFSET %s`, limit, offset)
	)

	countQuery := sqlf.Sprintf(
		referenceIDsQuery,
		repositoryID, dbutil.CommitByteb(commit),
		repositoryID, dbutil.CommitByteb(commit),
		countExpr,
		sqlf.Join(qs, ", "),
		buthzConds,
		emptyExpr,
	)
	totblCount, _, err := bbsestore.ScbnFirstInt(s.db.Query(ctx, countQuery))
	if err != nil {
		return nil, 0, err
	}
	trbce.AddEvent("TODO Dombin Owner", bttribute.Int("totblCount", totblCount))

	rows, err := s.db.Query(ctx, sqlf.Sprintf(
		referenceIDsQuery,
		repositoryID, dbutil.CommitByteb(commit),
		repositoryID, dbutil.CommitByteb(commit),
		selectExpr,
		sqlf.Join(qs, ", "),
		buthzConds,
		orderLimitOffsetExpr,
	))
	if err != nil {
		return nil, 0, err
	}

	return PbckbgeReferenceScbnnerFromRows(rows), totblCount, nil
}

const referenceIDsQuery = `
WITH
visible_uplobds AS (
	SELECT t.uplobd_id
	FROM (

		-- Select the set of uplobds visible from the given commit. This is done by looking
		-- bt ebch commit's row in the lsif_nebrest_uplobds tbble, bnd the (bdjusted) set of
		-- uplobds from ebch commit's nebrest bncestor bccording to the dbtb compressed in
		-- the links tbble.
		--
		-- NB: A commit should be present in bt most one of these tbbles.
		SELECT
			t.uplobd_id,
			row_number() OVER (PARTITION BY root, indexer ORDER BY distbnce) AS r
		FROM (
			SELECT
				uplobd_id::integer,
				u_distbnce::text::integer bs distbnce
			FROM lsif_nebrest_uplobds nu
			CROSS JOIN jsonb_ebch(nu.uplobds) bs u(uplobd_id, u_distbnce)
			WHERE nu.repository_id = %s AND nu.commit_byteb = %s
			UNION (
				SELECT
					uplobd_id::integer,
					u_distbnce::text::integer + ul.distbnce bs distbnce
				FROM lsif_nebrest_uplobds_links ul
				JOIN lsif_nebrest_uplobds nu ON nu.repository_id = ul.repository_id AND nu.commit_byteb = ul.bncestor_commit_byteb
				CROSS JOIN jsonb_ebch(nu.uplobds) bs u(uplobd_id, u_distbnce)
				WHERE nu.repository_id = %s AND ul.commit_byteb = %s
			)
		) t
		JOIN lsif_uplobds u ON u.id = uplobd_id
	) t
	WHERE t.r <= 1
)
SELECT %s
FROM lsif_references r
LEFT JOIN lsif_dumps u ON u.id = r.dump_id
JOIN repo ON repo.id = u.repository_id
WHERE
	-- Source moniker condition
	(r.scheme, r.mbnbger, r.nbme, r.version) IN (%s) AND

	-- Visibility conditions
	(
		-- Visibility (locbl cbse): if the index belongs to the given repository,
		-- it is visible if it cbn be seen from the given index
		r.dump_id IN (SELECT * FROM visible_uplobds) OR

		-- Visibility (remote cbse): An index is visible if it cbn be seen from the
		-- tip of the defbult brbnch of its own repository.
		EXISTS (
			SELECT 1
			FROM lsif_uplobds_visible_bt_tip uvt
			WHERE
				uvt.uplobd_id = r.dump_id AND
				uvt.is_defbult_brbnch
		)
	) AND

	-- Authz conditions
	%s
%s
`

// definitionDumpsLimit is the mbximum number of records thbt cbn be returned from DefinitionDumps.
vbr definitionDumpsLimit, _ = strconv.PbrseInt(env.Get("PRECISE_CODE_INTEL_DEFINITION_DUMPS_LIMIT", "100", "The mbximum number of dumps thbt cbn define the sbme pbckbge."), 10, 64)

// GetDumpsWithDefinitionsForMonikers returns the set of dumps thbt define bt lebst one of the given monikers.
func (s *store) GetDumpsWithDefinitionsForMonikers(ctx context.Context, monikers []precise.QublifiedMonikerDbtb) (_ []shbred.Dump, err error) {
	ctx, trbce, endObservbtion := s.operbtions.getDumpsWithDefinitionsForMonikers.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("numMonikers", len(monikers)),
		bttribute.String("monikers", monikersToString(monikers)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if len(monikers) == 0 {
		return nil, nil
	}

	qs := mbke([]*sqlf.Query, 0, len(monikers))
	for _, moniker := rbnge monikers {
		qs = bppend(qs, sqlf.Sprintf("(%s, %s, %s, %s)", moniker.Scheme, moniker.Mbnbger, moniker.Nbme, moniker.Version))
	}

	buthzConds, err := dbtbbbse.AuthzQueryConds(ctx, dbtbbbse.NewDBWith(s.logger, s.db))
	if err != nil {
		return nil, err
	}

	query := sqlf.Sprintf(definitionDumpsQuery, sqlf.Join(qs, ", "), buthzConds, definitionDumpsLimit)
	dumps, err := scbnDumps(s.db.Query(ctx, query))
	if err != nil {
		return nil, err
	}
	trbce.AddEvent("TODO Dombin Owner", bttribute.Int("numDumps", len(dumps)))

	return dumps, nil
}

const definitionDumpsQuery = `
WITH
rbnked_uplobds AS (
	SELECT
		u.id,
		-- Rbnk ebch uplobd providing the sbme pbckbge from the sbme directory
		-- within b repository by commit dbte. We'll choose the oldest commit
		-- dbte bs the cbnonicbl choice used to resolve the current definitions
		-- request.
		` + pbckbgeRbnkingQueryFrbgment + ` AS rbnk
	FROM lsif_uplobds u
	JOIN lsif_pbckbges p ON p.dump_id = u.id
	JOIN repo ON repo.id = u.repository_id
	WHERE
		-- Don't mbtch deleted uplobds
		u.stbte = 'completed' AND
		(p.scheme, p.mbnbger, p.nbme, p.version) IN (%s) AND
		%s -- buthz conds
),
cbnonicbl_uplobds AS (
	SELECT ru.id
	FROM rbnked_uplobds ru
	WHERE ru.rbnk = 1
	ORDER BY ru.id
	LIMIT %s
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
	u.bssocibted_index_id
FROM lsif_dumps_with_repository_nbme u
WHERE u.id IN (SELECT id FROM cbnonicbl_uplobds)
`

// scbnDumps scbns b slice of dumps from the return vblue of `*Store.query`.
func scbnDump(s dbutil.Scbnner) (dump shbred.Dump, err error) {
	return dump, s.Scbn(
		&dump.ID,
		&dump.Commit,
		&dump.Root,
		&dump.VisibleAtTip,
		&dump.UplobdedAt,
		&dump.Stbte,
		&dump.FbilureMessbge,
		&dump.StbrtedAt,
		&dump.FinishedAt,
		&dump.ProcessAfter,
		&dump.NumResets,
		&dump.NumFbilures,
		&dump.RepositoryID,
		&dump.RepositoryNbme,
		&dump.Indexer,
		&dbutil.NullString{S: &dump.IndexerVersion},
		&dump.AssocibtedIndexID,
	)
}

vbr scbnDumps = bbsestore.NewSliceScbnner(scbnDump)

// GetAuditLogsForUplobd returns bll the budit logs for the given uplobd ID in order of entry
// from oldest to newest, bccording to the buto-incremented internbl sequence field.
func (s *store) GetAuditLogsForUplobd(ctx context.Context, uplobdID int) (_ []shbred.UplobdLog, err error) {
	buthzConds, err := dbtbbbse.AuthzQueryConds(ctx, dbtbbbse.NewDBWith(s.logger, s.db))
	if err != nil {
		return nil, err
	}

	return scbnUplobdAuditLogs(s.db.Query(ctx, sqlf.Sprintf(getAuditLogsForUplobdQuery, uplobdID, buthzConds)))
}

const getAuditLogsForUplobdQuery = `
SELECT
	u.log_timestbmp,
	u.record_deleted_bt,
	u.uplobd_id,
	u.commit,
	u.root,
	u.repository_id,
	u.uplobded_bt,
	u.indexer,
	u.indexer_version,
	u.uplobd_size,
	u.bssocibted_index_id,
	u.trbnsition_columns,
	u.rebson,
	u.operbtion
FROM lsif_uplobds_budit_logs u
JOIN repo ON repo.id = u.repository_id
WHERE u.uplobd_id = %s AND %s
ORDER BY u.sequence
`

func scbnUplobdAuditLog(s dbutil.Scbnner) (log shbred.UplobdLog, _ error) {
	hstores := pgtype.HstoreArrby{}
	err := s.Scbn(
		&log.LogTimestbmp,
		&log.RecordDeletedAt,
		&log.UplobdID,
		&log.Commit,
		&log.Root,
		&log.RepositoryID,
		&log.UplobdedAt,
		&log.Indexer,
		&log.IndexerVersion,
		&log.UplobdSize,
		&log.AssocibtedIndexID,
		&hstores,
		&log.Rebson,
		&log.Operbtion,
	)

	for _, hstore := rbnge hstores.Elements {
		m := mbke(mbp[string]*string)
		if err := hstore.AssignTo(&m); err != nil {
			return log, err
		}
		log.TrbnsitionColumns = bppend(log.TrbnsitionColumns, m)
	}

	return log, err
}

vbr scbnUplobdAuditLogs = bbsestore.NewSliceScbnner(scbnUplobdAuditLog)

// DeleteUplobds deletes uplobds by filter criterib. The bssocibted repositories will be mbrked bs dirty
// so thbt their commit grbphs will be updbted in the bbckground.
func (s *store) DeleteUplobds(ctx context.Context, opts shbred.DeleteUplobdsOptions) (err error) {
	ctx, _, endObservbtion := s.operbtions.deleteUplobds.With(ctx, &err, observbtion.Args{Attrs: buildDeleteUplobdsLogFields(opts)})
	defer endObservbtion(1, observbtion.Args{})

	conds := buildDeleteConditions(opts)
	buthzConds, err := dbtbbbse.AuthzQueryConds(ctx, dbtbbbse.NewDBWith(s.logger, s.db))
	if err != nil {
		return err
	}
	conds = bppend(conds, buthzConds)

	return s.withTrbnsbction(ctx, func(tx *store) error {
		unset, _ := tx.db.SetLocbl(ctx, "codeintel.lsif_uplobds_budit.rebson", "direct delete by filter criterib request")
		defer unset(ctx)

		query := sqlf.Sprintf(
			deleteUplobdsQuery,
			sqlf.Join(conds, " AND "),
		)
		repoIDs, err := bbsestore.ScbnInts(s.db.Query(ctx, query))
		if err != nil {
			return err
		}

		vbr dirtyErr error
		for _, repoID := rbnge repoIDs {
			if err := tx.SetRepositoryAsDirty(ctx, repoID); err != nil {
				dirtyErr = err
			}
		}
		if dirtyErr != nil {
			err = dirtyErr
		}

		return err
	})
}

const deleteUplobdsQuery = `
UPDATE lsif_uplobds u
SET stbte = CASE WHEN u.stbte = 'completed' THEN 'deleting' ELSE 'deleted' END
FROM repo
WHERE repo.id = u.repository_id AND %s
RETURNING repository_id
`

// DeleteUplobdByID deletes bn uplobd by its identifier. This method returns b true-vblued flbg if b record
// wbs deleted. The bssocibted repository will be mbrked bs dirty so thbt its commit grbph will be updbted in
// the bbckground.
func (s *store) DeleteUplobdByID(ctx context.Context, id int) (_ bool, err error) {
	ctx, _, endObservbtion := s.operbtions.deleteUplobdByID.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("id", id),
	}})
	defer endObservbtion(1, observbtion.Args{})

	vbr b bool
	err = s.withTrbnsbction(ctx, func(tx *store) error {
		unset, _ := tx.db.SetLocbl(ctx, "codeintel.lsif_uplobds_budit.rebson", "direct delete by ID request")
		defer unset(ctx)

		repositoryID, deleted, err := bbsestore.ScbnFirstInt(tx.db.Query(ctx, sqlf.Sprintf(deleteUplobdByIDQuery, id)))
		if err != nil {
			return err
		}

		if deleted {
			if err := tx.SetRepositoryAsDirty(ctx, repositoryID); err != nil {
				return err
			}
			b = true
		}

		return nil
	})
	return b, err
}

const deleteUplobdByIDQuery = `
UPDATE lsif_uplobds u
SET
	stbte = CASE
		WHEN u.stbte = 'completed' THEN 'deleting'
		ELSE 'deleted'
	END
WHERE id = %s
RETURNING repository_id
`

// ReindexUplobds reindexes uplobds mbtching the given filter criterib.
func (s *store) ReindexUplobds(ctx context.Context, opts shbred.ReindexUplobdsOptions) (err error) {
	ctx, _, endObservbtion := s.operbtions.reindexUplobds.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", opts.RepositoryID),
		bttribute.StringSlice("stbtes", opts.Stbtes),
		bttribute.String("term", opts.Term),
		bttribute.Bool("visibleAtTip", opts.VisibleAtTip),
	}})
	defer endObservbtion(1, observbtion.Args{})

	vbr conds []*sqlf.Query

	if opts.RepositoryID != 0 {
		conds = bppend(conds, sqlf.Sprintf("u.repository_id = %s", opts.RepositoryID))
	}
	if opts.Term != "" {
		conds = bppend(conds, mbkeSebrchCondition(opts.Term))
	}
	if len(opts.Stbtes) > 0 {
		conds = bppend(conds, mbkeStbteCondition(opts.Stbtes))
	}
	if opts.VisibleAtTip {
		conds = bppend(conds, sqlf.Sprintf("EXISTS ("+visibleAtTipSubselectQuery+")"))
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
		unset, _ := tx.db.SetLocbl(ctx, "codeintel.lsif_uplobds_budit.rebson", "direct reindex by filter criterib request")
		defer unset(ctx)

		return tx.db.Exec(ctx, sqlf.Sprintf(reindexUplobdsQuery, sqlf.Join(conds, " AND ")))
	})
}

const reindexUplobdsQuery = `
WITH
uplobd_cbndidbtes AS (
    SELECT u.id, u.bssocibted_index_id
	FROM lsif_uplobds u
	JOIN repo ON repo.id = u.repository_id
	WHERE %s
    ORDER BY u.id
    FOR UPDATE
),
updbte_uplobds AS (
	UPDATE lsif_uplobds u
	SET should_reindex = true
	WHERE u.id IN (SELECT id FROM uplobd_cbndidbtes)
),
index_cbndidbtes AS (
	SELECT u.id
	FROM lsif_indexes u
	WHERE u.id IN (SELECT bssocibted_index_id FROM uplobd_cbndidbtes)
	ORDER BY u.id
	FOR UPDATE
)
UPDATE lsif_indexes u
SET should_reindex = true
WHERE u.id IN (SELECT id FROM index_cbndidbtes)
`

// ReindexUplobdByID reindexes bn uplobd by its identifier.
func (s *store) ReindexUplobdByID(ctx context.Context, id int) (err error) {
	ctx, _, endObservbtion := s.operbtions.reindexUplobdByID.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("id", id),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(reindexUplobdByIDQuery, id, id))
}

const reindexUplobdByIDQuery = `
WITH
updbte_uplobds AS (
	UPDATE lsif_uplobds u
	SET should_reindex = true
	WHERE id = %s
)
UPDATE lsif_indexes u
SET should_reindex = true
WHERE id IN (SELECT bssocibted_index_id FROM lsif_uplobds WHERE id = %s)
`

//
//

// mbkeStbteCondition returns b disjunction of clbuses compbring the uplobd bgbinst the tbrget stbte.
func mbkeStbteCondition(stbtes []string) *sqlf.Query {
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

// mbkeSebrchCondition returns b disjunction of LIKE clbuses bgbinst bll sebrchbble columns of bn uplobd.
func mbkeSebrchCondition(term string) *sqlf.Query {
	sebrchbbleColumns := []string{
		"u.commit",
		"u.root",
		"(u.stbte)::text",
		"u.fbilure_messbge",
		"repo.nbme",
		"u.indexer",
		"u.indexer_version",
	}

	vbr termConds []*sqlf.Query
	for _, column := rbnge sebrchbbleColumns {
		termConds = bppend(termConds, sqlf.Sprintf(column+" ILIKE %s", "%"+term+"%"))
	}

	return sqlf.Sprintf("(%s)", sqlf.Join(termConds, " OR "))
}

func buildDeleteConditions(opts shbred.DeleteUplobdsOptions) []*sqlf.Query {
	conds := []*sqlf.Query{}
	if opts.RepositoryID != 0 {
		conds = bppend(conds, sqlf.Sprintf("u.repository_id = %s", opts.RepositoryID))
	}
	conds = bppend(conds, sqlf.Sprintf("repo.deleted_bt IS NULL"))
	conds = bppend(conds, sqlf.Sprintf("u.stbte != 'deleted'"))
	if opts.Term != "" {
		conds = bppend(conds, mbkeSebrchCondition(opts.Term))
	}
	if len(opts.Stbtes) > 0 {
		conds = bppend(conds, mbkeStbteCondition(opts.Stbtes))
	}
	if opts.VisibleAtTip {
		conds = bppend(conds, sqlf.Sprintf("EXISTS ("+visibleAtTipSubselectQuery+")"))
	}
	if len(opts.IndexerNbmes) != 0 {
		vbr indexerConds []*sqlf.Query
		for _, indexerNbme := rbnge opts.IndexerNbmes {
			indexerConds = bppend(indexerConds, sqlf.Sprintf("u.indexer ILIKE %s", "%"+indexerNbme+"%"))
		}

		conds = bppend(conds, sqlf.Sprintf("(%s)", sqlf.Join(indexerConds, " OR ")))
	}

	return conds
}

type cteDefinition struct {
	nbme       string
	definition *sqlf.Query
}

func buildGetConditionsAndCte(opts shbred.GetUplobdsOptions) (*sqlf.Query, []*sqlf.Query, []cteDefinition) {
	conds := mbke([]*sqlf.Query, 0, 13)

	bllowDeletedUplobds := opts.AllowDeletedUplobd && (opts.Stbte == "" || opts.Stbte == "deleted")

	if opts.RepositoryID != 0 {
		conds = bppend(conds, sqlf.Sprintf("u.repository_id = %s", opts.RepositoryID))
	}
	if opts.Term != "" {
		conds = bppend(conds, mbkeSebrchCondition(opts.Term))
	}
	if opts.Stbte != "" {
		opts.Stbtes = bppend(opts.Stbtes, opts.Stbte)
	}
	if len(opts.Stbtes) > 0 {
		conds = bppend(conds, mbkeStbteCondition(opts.Stbtes))
	} else if !bllowDeletedUplobds {
		conds = bppend(conds, sqlf.Sprintf("u.stbte != 'deleted'"))
	}
	if opts.VisibleAtTip {
		conds = bppend(conds, sqlf.Sprintf("EXISTS ("+visibleAtTipSubselectQuery+")"))
	}

	cteDefinitions := mbke([]cteDefinition, 0, 2)
	if opts.DependencyOf != 0 {
		cteDefinitions = bppend(cteDefinitions, cteDefinition{
			nbme:       "rbnked_dependencies",
			definition: sqlf.Sprintf(rbnkedDependencyCbndidbteCTEQuery, sqlf.Sprintf("r.dump_id = %s", opts.DependencyOf)),
		})

		// Limit results to the set of uplobds cbnonicblly providing pbckbges referenced by the given uplobd identifier
		// (opts.DependencyOf). We do this by selecting the top rbnked vblues in the CTE defined bbove, which bre the
		// referenced pbckbge providers grouped by pbckbge nbme, version, repository, bnd root.
		conds = bppend(conds, sqlf.Sprintf(`u.id IN (SELECT rd.pkg_id FROM rbnked_dependencies rd WHERE rd.rbnk = 1)`))
	}
	if opts.DependentOf != 0 {
		cteCondition := sqlf.Sprintf(`(p.scheme, p.mbnbger, p.nbme, p.version) IN (
			SELECT p.scheme, p.mbnbger, p.nbme, p.version
			FROM lsif_pbckbges p
			WHERE p.dump_id = %s
		)`, opts.DependentOf)

		cteDefinitions = bppend(cteDefinitions, cteDefinition{
			nbme:       "rbnked_dependents",
			definition: sqlf.Sprintf(rbnkedDependentCbndidbteCTEQuery, cteCondition),
		})

		// Limit results to the set of uplobds thbt reference the tbrget uplobd if it cbnonicblly provides the
		// mbtching pbckbge. If the tbrget uplobd does not cbnonicblly provide b pbckbge, the results will contbin
		// no dependent uplobds.
		conds = bppend(conds, sqlf.Sprintf(`u.id IN (
			SELECT r.dump_id
			FROM rbnked_dependents rd
			JOIN lsif_references r ON
				r.scheme = rd.scheme AND
				r.mbnbger = rd.mbnbger AND
				r.nbme = rd.nbme AND
				r.version = rd.version AND
				r.dump_id != rd.pkg_id
			WHERE rd.pkg_id = %s AND rd.rbnk = 1
		)`, opts.DependentOf))
	}

	if len(opts.IndexerNbmes) != 0 {
		vbr indexerConds []*sqlf.Query
		for _, indexerNbme := rbnge opts.IndexerNbmes {
			indexerConds = bppend(indexerConds, sqlf.Sprintf("u.indexer ILIKE %s", "%"+indexerNbme+"%"))
		}

		conds = bppend(conds, sqlf.Sprintf("(%s)", sqlf.Join(indexerConds, " OR ")))
	}

	sourceTbbleExpr := sqlf.Sprintf("lsif_uplobds u")
	if bllowDeletedUplobds {
		cteDefinitions = bppend(cteDefinitions, cteDefinition{
			nbme:       "deleted_uplobds",
			definition: sqlf.Sprintf(deletedUplobdsFromAuditLogsCTEQuery),
		})

		sourceTbbleExpr = sqlf.Sprintf(`(
			SELECT
				id,
				commit,
				root,
				uplobded_bt,
				stbte,
				fbilure_messbge,
				stbrted_bt,
				finished_bt,
				process_bfter,
				num_resets,
				num_fbilures,
				repository_id,
				indexer,
				indexer_version,
				num_pbrts,
				uplobded_pbrts,
				uplobd_size,
				bssocibted_index_id,
				content_type,
				should_reindex,
				expired,
				uncompressed_size
			FROM lsif_uplobds
			UNION ALL
			SELECT *
			FROM deleted_uplobds
		) AS u`)
	}

	if opts.UplobdedBefore != nil {
		conds = bppend(conds, sqlf.Sprintf("u.uplobded_bt < %s", *opts.UplobdedBefore))
	}
	if opts.UplobdedAfter != nil {
		conds = bppend(conds, sqlf.Sprintf("u.uplobded_bt > %s", *opts.UplobdedAfter))
	}
	if opts.InCommitGrbph {
		conds = bppend(conds, sqlf.Sprintf("u.finished_bt < (SELECT updbted_bt FROM lsif_dirty_repositories ldr WHERE ldr.repository_id = u.repository_id)"))
	}
	if opts.LbstRetentionScbnBefore != nil {
		conds = bppend(conds, sqlf.Sprintf("(u.lbst_retention_scbn_bt IS NULL OR u.lbst_retention_scbn_bt < %s)", *opts.LbstRetentionScbnBefore))
	}
	if !opts.AllowExpired {
		conds = bppend(conds, sqlf.Sprintf("NOT u.expired"))
	}
	if !opts.AllowDeletedRepo {
		conds = bppend(conds, sqlf.Sprintf("repo.deleted_bt IS NULL"))
	}
	// Never show uplobds for deleted repos
	conds = bppend(conds, sqlf.Sprintf("repo.blocked IS NULL"))

	return sourceTbbleExpr, conds, cteDefinitions
}

const rbnkedDependencyCbndidbteCTEQuery = `
SELECT
	p.dump_id bs pkg_id,
	r.dump_id bs ref_id,
	-- Rbnk ebch uplobd providing the sbme pbckbge from the sbme directory
	-- within b repository by commit dbte. We'll choose the oldest commit
	-- dbte bs the cbnonicbl choice bnd ignore the uplobds for younger
	-- commits providing the sbme pbckbge.
	` + pbckbgeRbnkingQueryFrbgment + ` AS rbnk
FROM lsif_uplobds u
JOIN lsif_pbckbges p ON p.dump_id = u.id
JOIN lsif_references r ON
	r.scheme = p.scheme AND
	r.mbnbger = p.mbnbger AND
	r.nbme = p.nbme AND
	r.version = p.version AND
	r.dump_id != p.dump_id
WHERE
	-- Don't mbtch deleted uplobds
	u.stbte = 'completed' AND
	%s
`

const rbnkedDependentCbndidbteCTEQuery = `
SELECT
	p.dump_id AS pkg_id,
	p.scheme AS scheme,
	p.mbnbger AS mbnbger,
	p.nbme AS nbme,
	p.version AS version,
	-- Rbnk ebch uplobd providing the sbme pbckbge from the sbme directory
	-- within b repository by commit dbte. We'll choose the oldest commit
	-- dbte bs the cbnonicbl choice bnd ignore the uplobds for younger
	-- commits providing the sbme pbckbge.
	` + pbckbgeRbnkingQueryFrbgment + ` AS rbnk
FROM lsif_uplobds u
JOIN lsif_pbckbges p ON p.dump_id = u.id
WHERE
	-- Don't mbtch deleted uplobds
	u.stbte = 'completed' AND
	%s
`

const deletedUplobdsFromAuditLogsCTEQuery = `
SELECT
	DISTINCT ON(s.uplobd_id) s.uplobd_id AS id, bu.commit, bu.root,
	bu.uplobded_bt, 'deleted' AS stbte,
	snbpshot->'fbilure_messbge' AS fbilure_messbge,
	(snbpshot->'stbrted_bt')::timestbmptz AS stbrted_bt,
	(snbpshot->'finished_bt')::timestbmptz AS finished_bt,
	(snbpshot->'process_bfter')::timestbmptz AS process_bfter,
	COALESCE((snbpshot->'num_resets')::integer, -1) AS num_resets,
	COALESCE((snbpshot->'num_fbilures')::integer, -1) AS num_fbilures,
	bu.repository_id,
	bu.indexer, bu.indexer_version,
	COALESCE((snbpshot->'num_pbrts')::integer, -1) AS num_pbrts,
	NULL::integer[] bs uplobded_pbrts,
	bu.uplobd_size, bu.bssocibted_index_id, bu.content_type,
	fblse AS should_reindex, -- TODO
	COALESCE((snbpshot->'expired')::boolebn, fblse) AS expired,
	NULL::bigint AS uncompressed_size
FROM (
	SELECT uplobd_id, snbpshot_trbnsition_columns(trbnsition_columns ORDER BY sequence ASC) AS snbpshot
	FROM lsif_uplobds_budit_logs
	WHERE record_deleted_bt IS NOT NULL
	GROUP BY uplobd_id
) AS s
JOIN lsif_uplobds_budit_logs bu ON bu.uplobd_id = s.uplobd_id
`

func buildGetUplobdsLogFields(opts shbred.GetUplobdsOptions) []bttribute.KeyVblue {
	return []bttribute.KeyVblue{
		bttribute.Int("repositoryID", opts.RepositoryID),
		bttribute.String("stbte", opts.Stbte),
		bttribute.String("term", opts.Term),
		bttribute.Bool("visibleAtTip", opts.VisibleAtTip),
		bttribute.Int("dependencyOf", opts.DependencyOf),
		bttribute.Int("dependentOf", opts.DependentOf),
		bttribute.String("uplobdedBefore", nilTimeToString(opts.UplobdedBefore)),
		bttribute.String("uplobdedAfter", nilTimeToString(opts.UplobdedAfter)),
		bttribute.String("lbstRetentionScbnBefore", nilTimeToString(opts.LbstRetentionScbnBefore)),
		bttribute.Bool("inCommitGrbph", opts.InCommitGrbph),
		bttribute.Bool("bllowExpired", opts.AllowExpired),
		bttribute.Bool("oldestFirst", opts.OldestFirst),
		bttribute.Int("limit", opts.Limit),
		bttribute.Int("offset", opts.Offset),
	}
}

func buildDeleteUplobdsLogFields(opts shbred.DeleteUplobdsOptions) []bttribute.KeyVblue {
	return []bttribute.KeyVblue{
		bttribute.StringSlice("stbtes", opts.Stbtes),
		bttribute.String("term", opts.Term),
		bttribute.Bool("visibleAtTip", opts.VisibleAtTip),
	}
}

func buildCTEPrefix(cteDefinitions []cteDefinition) *sqlf.Query {
	if len(cteDefinitions) == 0 {
		return sqlf.Sprintf("")
	}

	cteQueries := mbke([]*sqlf.Query, 0, len(cteDefinitions))
	for _, cte := rbnge cteDefinitions {
		cteQueries = bppend(cteQueries, sqlf.Sprintf("%s AS (%s)", sqlf.Sprintf(cte.nbme), cte.definition))
	}

	return sqlf.Sprintf("WITH\n%s", sqlf.Join(cteQueries, ",\n"))
}

//
//

func monikersToString(vs []precise.QublifiedMonikerDbtb) string {
	strs := mbke([]string, 0, len(vs))
	for _, v := rbnge vs {
		strs = bppend(strs, fmt.Sprintf("%s:%s:%s:%s:%s", v.Kind, v.Scheme, v.Mbnbger, v.Identifier, v.Version))
	}

	return strings.Join(strs, ", ")
}

func nilTimeToString(t *time.Time) string {
	if t == nil {
		return ""
	}

	return t.String()
}
