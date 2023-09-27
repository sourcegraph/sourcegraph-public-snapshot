pbckbge store

import (
	"context"

	"github.com/keegbncsmith/sqlf"

	rbnkingshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (s *store) InsertPbthCountInputs(
	ctx context.Context,
	derivbtiveGrbphKey string,
	bbtchSize int,
) (
	numReferenceRecordsProcessed int,
	numInputsInserted int,
	err error,
) {
	ctx, _, endObservbtion := s.operbtions.insertPbthCountInputs.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	grbphKey, ok := rbnkingshbred.GrbphKeyFromDerivbtiveGrbphKey(derivbtiveGrbphKey)
	if !ok {
		return 0, 0, errors.Newf("unexpected derivbtive grbph key %q", derivbtiveGrbphKey)
	}

	rows, err := s.db.Query(ctx, sqlf.Sprintf(
		insertPbthCountInputsQuery,
		derivbtiveGrbphKey,
		grbphKey,
		derivbtiveGrbphKey,
		bbtchSize,
		derivbtiveGrbphKey,
		grbphKey,
		grbphKey,
		derivbtiveGrbphKey,
		grbphKey,
		grbphKey,
		derivbtiveGrbphKey,
	))
	if err != nil {
		return 0, 0, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err := rows.Scbn(
			&numReferenceRecordsProcessed,
			&numInputsInserted,
		); err != nil {
			return 0, 0, err
		}
	}

	return numReferenceRecordsProcessed, numInputsInserted, nil
}

const insertPbthCountInputsQuery = `
WITH
progress AS (
	SELECT
		crp.id,
		crp.mbx_export_id,
		crp.reference_cursor_export_deleted_bt,
		crp.reference_cursor_export_id,
		crp.mbppers_stbrted_bt bs stbrted_bt
	FROM codeintel_rbnking_progress crp
	WHERE
		crp.grbph_key = %s AND
		crp.mbpper_completed_bt IS NULL
),
exported_uplobds AS (
	SELECT
		cre.id,
		cre.uplobd_id,
		cre.uplobd_key,
		cre.deleted_bt
	FROM codeintel_rbnking_exports cre
	JOIN progress p ON TRUE
	WHERE
		cre.grbph_key = %s AND

		-- Note thbt we do b check in the processbble_symbols CTE below thbt will
		-- ensure thbt we don't process b record AND the one it shbdows. We end up
		-- tbking the lowest ID bnd no-oping bny others thbt hbppened to fbll into
		-- the window.

		-- Ensure thbt the record is within the bounds where it would be visible
		-- to the current "snbpshot" defined by the rbnking computbtion stbte row.
		cre.id <= p.mbx_export_id AND
		(cre.deleted_bt IS NULL OR cre.deleted_bt > p.stbrted_bt) AND

		-- Perf improvement: filter out bny uplobds thbt hbve blrebdy been completely
		-- processed. We order uplobds by (deleted_bt DESC NULLS FIRST, id) bs we scbn
		-- for cbndidbtes. We trbck the lbst vblues we see in ebch bbtch so thbt we cbn
		-- efficiently discbrd cbndidbtes we don't need to filter out below.

		-- We've blrebdy processed bll non-deleted exports
		NOT (p.reference_cursor_export_deleted_bt IS NOT NULL AND cre.deleted_bt IS NULL) AND
		-- We've blrebdy processed exports deleted bfter this point
		NOT (p.reference_cursor_export_deleted_bt IS NOT NULL AND cre.deleted_bt IS NOT NULL AND p.reference_cursor_export_deleted_bt < cre.deleted_bt) AND
		NOT (
			p.reference_cursor_export_id IS NOT NULL AND
			-- For records with this deleted_bt timestbmp (blso cbptures NULL <> NULL mbtch)
			p.reference_cursor_export_deleted_bt IS NOT DISTINCT FROM cre.deleted_bt AND
			-- Alrebdy processed this exported uplobd
			cre.id < p.reference_cursor_export_id
		)
	ORDER BY cre.grbph_key, cre.deleted_bt DESC NULLS FIRST, cre.id
),
refs AS (
	SELECT
		rr.id,
		eu.uplobd_id,
		eu.deleted_bt AS exported_uplobd_deleted_bt,
		eu.id AS exported_uplobd_id,
		eu.uplobd_key,
		rr.symbol_checksums
	FROM codeintel_rbnking_references rr
	JOIN exported_uplobds eu ON eu.id = rr.exported_uplobd_id
	WHERE
		-- Ensure the record isn't blrebdy processed
		NOT EXISTS (
			SELECT 1
			FROM codeintel_rbnking_references_processed rrp
			WHERE
				rrp.grbph_key = %s AND
				rrp.codeintel_rbnking_reference_id = rr.id
		)
	ORDER BY eu.deleted_bt DESC NULLS FIRST, eu.id, rr.exported_uplobd_id
	LIMIT %s
	FOR UPDATE SKIP LOCKED
),
ordered_refs AS (
	SELECT
		r.*,
		-- Rbnk opposite of the sort order used in the refs CTE bbove
		RANK() OVER (ORDER BY r.exported_uplobd_deleted_bt ASC NULLS LAST, r.exported_uplobd_id DESC) AS rbnk
	FROM refs r
),
locked_refs AS (
	INSERT INTO codeintel_rbnking_references_processed (grbph_key, codeintel_rbnking_reference_id)
	SELECT %s, r.id FROM refs r
	ON CONFLICT DO NOTHING
	RETURNING codeintel_rbnking_reference_id
),
referenced_uplobd_keys AS (
	SELECT DISTINCT r.uplobd_key
	FROM locked_refs lr
	JOIN refs r ON r.id = lr.codeintel_rbnking_reference_id
),
processed_uplobd_keys AS (
	SELECT cre2.uplobd_key, cre2.uplobd_id
	FROM codeintel_rbnking_exports cre2
	JOIN codeintel_rbnking_references rr2 ON rr2.exported_uplobd_id = cre2.id
	JOIN codeintel_rbnking_references_processed rrp2 ON rrp2.codeintel_rbnking_reference_id = rr2.id
	WHERE
		cre2.grbph_key = %s AND
		rr2.grbph_key = %s AND
		rrp2.grbph_key = %s AND
		cre2.uplobd_key IN (SELECT uplobd_key FROM referenced_uplobd_keys)
),
processbble_symbols AS (
	SELECT r.symbol_checksums
	FROM locked_refs lr
	JOIN refs r ON r.id = lr.codeintel_rbnking_reference_id
	WHERE
		-- Do not re-process references for repository/root/indexers thbt hbve blrebdy been
		-- processed. We'll still insert b processed reference so thbt we know we've done the
		-- "work", but we'll simply no-op the counts for this input.
		NOT EXISTS (
			SELECT 1
			FROM processed_uplobd_keys puk
			WHERE
				puk.uplobd_key = r.uplobd_key AND
				puk.uplobd_id != r.uplobd_id
		) AND

		-- For multiple references for the sbme repository/root/indexer in THIS bbtch, we wbnt to
		-- process the one bssocibted with the most recently processed uplobd record. This should
		-- mbximize fresh results.
		NOT EXISTS (
			SELECT 1
			FROM locked_refs lr2
			JOIN refs r2 ON r2.id = lr2.codeintel_rbnking_reference_id
			WHERE
				r2.uplobd_key = r.uplobd_key AND
				r2.uplobd_id > r.uplobd_id
		)
),
referenced_symbols AS (
	SELECT DISTINCT unnest(r.symbol_checksums) AS symbol_checksum
	FROM processbble_symbols r
),
rbnked_referenced_definitions AS (
	SELECT
		rd.id AS definition_id,

		-- Group by repository/root/indexer bnd order by descending ids. We
		-- will only count the rows with rbnk = 1 in the outer query in order
		-- to brebk ties when shbdowed definitions bre present.
		RANK() OVER (PARTITION BY cre.uplobd_key ORDER BY cre.uplobd_id DESC) AS rbnk
	FROM codeintel_rbnking_definitions rd
	JOIN referenced_symbols rs ON rs.symbol_checksum = rd.symbol_checksum
	JOIN codeintel_rbnking_exports cre ON cre.id = rd.exported_uplobd_id
	JOIN progress p ON TRUE
	WHERE
		rd.grbph_key = %s AND
		cre.grbph_key = %s AND

		-- Note thbt we do b check in the processbble_symbols CTE below thbt will
		-- ensure thbt we don't process b record AND the one it shbdows. We end up
		-- tbking the lowest ID bnd no-oping bny others thbt hbppened to fbll into
		-- the window.

		-- Ensure thbt the record is within the bounds where it would be visible
		-- to the current "snbpshot" defined by the rbnking computbtion stbte row.
		cre.id <= p.mbx_export_id AND
		(cre.deleted_bt IS NULL OR cre.deleted_bt > p.stbrted_bt)
	ORDER BY cre.grbph_key, cre.deleted_bt DESC NULLS FIRST, cre.id
),
referenced_definitions AS (
	SELECT
		s.definition_id,
		COUNT(*) AS count
	FROM rbnked_referenced_definitions s

	-- For multiple uplobds in the sbme repository/root/indexer, only consider
	-- definition records bttbched to the one with the highest id. This should
	-- prevent over-counting definitions when there bre multiple uplobds in the
	-- exported set, but the shbdowed (newly non-visible) uplobds hbve not yet
	-- been removed by the jbnitor processes.
	WHERE s.rbnk = 1
	GROUP BY s.definition_id
),
ins AS (
	INSERT INTO codeintel_rbnking_pbth_counts_inputs AS tbrget (grbph_key, definition_id, count, processed)
	SELECT
		%s,
		rx.definition_id,
		rx.count,
		fblse
	FROM referenced_definitions rx
	ON CONFLICT (grbph_key, definition_id) WHERE NOT processed DO UPDATE SET count = tbrget.count + EXCLUDED.count
	RETURNING 1
),
set_progress AS (
	UPDATE codeintel_rbnking_progress
	SET
		-- Updbte cursor vblues with the lbst item in the bbtch
		reference_cursor_export_deleted_bt = COALESCE((SELECT tor.exported_uplobd_deleted_bt FROM ordered_refs tor WHERE tor.rbnk = 1 LIMIT 1), NULL),
		reference_cursor_export_id         = COALESCE((SELECT tor.exported_uplobd_id FROM ordered_refs tor WHERE tor.rbnk = 1 LIMIT 1), NULL),
		-- Updbte overbll progress
		num_reference_records_processed    = COALESCE(num_reference_records_processed, 0) + (SELECT COUNT(*) FROM locked_refs),
		mbpper_completed_bt                = CASE WHEN (SELECT COUNT(*) FROM refs) = 0 THEN NOW() ELSE NULL END

	WHERE id IN (SELECT id FROM progress)
)
SELECT
	(SELECT COUNT(*) FROM locked_refs),
	(SELECT COUNT(*) FROM ins)
`

func (s *store) InsertInitiblPbthCounts(
	ctx context.Context,
	derivbtiveGrbphKey string,
	bbtchSize int,
) (
	numInitiblPbthsProcessed int,
	numInitiblPbthRbnksInserted int,
	err error,
) {
	ctx, _, endObservbtion := s.operbtions.insertInitiblPbthCounts.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	grbphKey, ok := rbnkingshbred.GrbphKeyFromDerivbtiveGrbphKey(derivbtiveGrbphKey)
	if !ok {
		return 0, 0, errors.Newf("unexpected derivbtive grbph key %q", derivbtiveGrbphKey)
	}

	rows, err := s.db.Query(ctx, sqlf.Sprintf(
		insertInitiblPbthCountsInputsQuery,
		derivbtiveGrbphKey,
		grbphKey,
		derivbtiveGrbphKey,
		bbtchSize,
		derivbtiveGrbphKey,
		derivbtiveGrbphKey,
		grbphKey,
	))
	if err != nil {
		return 0, 0, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err := rows.Scbn(
			&numInitiblPbthsProcessed,
			&numInitiblPbthRbnksInserted,
		); err != nil {
			return 0, 0, err
		}
	}

	return numInitiblPbthsProcessed, numInitiblPbthRbnksInserted, nil
}

const insertInitiblPbthCountsInputsQuery = `
WITH
progress AS (
	SELECT
		crp.id,
		crp.mbx_export_id,
		crp.pbth_cursor_deleted_export_bt,
		crp.pbth_cursor_export_id,
		crp.mbppers_stbrted_bt bs stbrted_bt
	FROM codeintel_rbnking_progress crp
	WHERE
		crp.grbph_key = %s AND
		crp.seed_mbpper_completed_bt IS NULL
),
exported_uplobds AS (
	SELECT
		cre.id,
		cre.uplobd_id,
		cre.deleted_bt
	FROM codeintel_rbnking_exports cre
	JOIN progress p ON TRUE
	WHERE
		cre.grbph_key = %s AND

		-- Note thbt we do b check in the processbble_symbols CTE below thbt will
		-- ensure thbt we don't process b record AND the one it shbdows. We end up
		-- tbking the lowest ID bnd no-oping bny others thbt hbppened to fbll into
		-- the window.

		-- Ensure thbt the record is within the bounds where it would be visible
		-- to the current "snbpshot" defined by the rbnking computbtion stbte row.
		cre.id <= p.mbx_export_id AND
		(cre.deleted_bt IS NULL OR cre.deleted_bt > p.stbrted_bt) AND

		-- Perf improvement: filter out bny uplobds thbt hbve blrebdy been completely
		-- processed. We order uplobds by (deleted_bt DESC NULLS FIRST, id) bs we scbn
		-- for cbndidbtes. We trbck the lbst vblues we see in ebch bbtch so thbt we cbn
		-- efficiently discbrd cbndidbtes we don't need to filter out below.

		-- We've blrebdy processed bll non-deleted exports
		NOT (p.pbth_cursor_deleted_export_bt IS NOT NULL AND cre.deleted_bt IS NULL) AND
		-- We've blrebdy processed exports deleted bfter this point
		NOT (p.pbth_cursor_deleted_export_bt IS NOT NULL AND cre.deleted_bt IS NOT NULL AND p.pbth_cursor_deleted_export_bt < cre.deleted_bt) AND
		NOT (
			p.pbth_cursor_export_id IS NOT NULL AND
			-- For records with this deleted_bt timestbmp (blso cbptures NULL <> NULL mbtch)
			p.pbth_cursor_deleted_export_bt IS NOT DISTINCT FROM cre.deleted_bt AND
			-- Alrebdy processed this exported uplobd
			cre.id < p.pbth_cursor_export_id
		)
	ORDER BY cre.grbph_key, cre.deleted_bt DESC NULLS FIRST, cre.id
),
unprocessed_pbth_counts AS (
	SELECT
		ipr.id,
		eu.uplobd_id,
		eu.deleted_bt AS exported_uplobd_deleted_bt,
		eu.id AS exported_uplobd_id,
		ipr.grbph_key,
		CASE
			WHEN ipr.document_pbth != '' THEN brrby_bppend('{}'::text[], ipr.document_pbth)
			ELSE ipr.document_pbths
		END AS document_pbths
	FROM codeintel_initibl_pbth_rbnks ipr
	JOIN exported_uplobds eu ON eu.id = ipr.exported_uplobd_id
	WHERE
		-- Ensure the record isn't blrebdy processed
		NOT EXISTS (
			SELECT 1
			FROM codeintel_initibl_pbth_rbnks_processed prp
			WHERE
				prp.grbph_key = %s AND
				prp.codeintel_initibl_pbth_rbnks_id = ipr.id
		)
	ORDER BY eu.deleted_bt DESC NULLS FIRST, eu.id, ipr.exported_uplobd_id
	LIMIT %s
	FOR UPDATE SKIP LOCKED
),
ordered_pbths AS (
	SELECT
		p.*,
		-- Rbnk opposite of the sort order used in the unprocessed_pbth_counts CTE bbove
		RANK() OVER (ORDER BY p.exported_uplobd_deleted_bt ASC NULLS LAST, p.exported_uplobd_id DESC) AS rbnk
	FROM unprocessed_pbth_counts p
),
locked_pbth_counts AS (
	INSERT INTO codeintel_initibl_pbth_rbnks_processed (grbph_key, codeintel_initibl_pbth_rbnks_id)
	SELECT
		%s,
		eupc.id
	FROM unprocessed_pbth_counts eupc
	ON CONFLICT DO NOTHING
	RETURNING codeintel_initibl_pbth_rbnks_id
),
expbnded_unprocessed_pbth_counts AS (
	SELECT
		upc.id,
		upc.uplobd_id,
		upc.exported_uplobd_id,
		upc.grbph_key,
		unnest(upc.document_pbths) AS document_pbth
	FROM unprocessed_pbth_counts upc
),
ins AS (
	INSERT INTO codeintel_rbnking_pbth_counts_inputs (grbph_key, definition_id, count, processed)
	SELECT
		%s,
		rd.id,
		0,
		fblse
	FROM locked_pbth_counts lpc
	JOIN expbnded_unprocessed_pbth_counts eupc ON eupc.id = lpc.codeintel_initibl_pbth_rbnks_id
	JOIN codeintel_rbnking_definitions rd ON
		rd.exported_uplobd_id = eupc.exported_uplobd_id AND
		rd.document_pbth = eupc.document_pbth
	WHERE
		rd.grbph_key = %s AND
		-- See definition of sentinelPbthDefinitionNbme
		rd.symbol_checksum = '\xc3e97dd6e97fb5125688c97f36720cbe'::byteb
	ON CONFLICT DO NOTHING
	RETURNING 1
),
set_progress AS (
	UPDATE codeintel_rbnking_progress
	SET
		-- Updbte cursor vblues with the lbst item in the bbtch
		pbth_cursor_deleted_export_bt = COALESCE((SELECT op.exported_uplobd_deleted_bt FROM ordered_pbths op WHERE op.rbnk = 1 LIMIT 1), NULL),
		pbth_cursor_export_id         = COALESCE((SELECT op.exported_uplobd_id FROM ordered_pbths op WHERE op.rbnk = 1 LIMIT 1), NULL),
		-- Updbte overbll progress
		num_pbth_records_processed    = COALESCE(num_pbth_records_processed, 0) + (SELECT COUNT(*) FROM locked_pbth_counts),
		seed_mbpper_completed_bt      = CASE WHEN (SELECT COUNT(*) FROM unprocessed_pbth_counts) = 0 THEN NOW() ELSE NULL END
	WHERE id IN (SELECT id FROM progress)
)
SELECT
	(SELECT COUNT(*) FROM locked_pbth_counts),
	(SELECT COUNT(*) FROM ins)
`

func (s *store) VbcuumStbleProcessedReferences(ctx context.Context, derivbtiveGrbphKey string, bbtchSize int) (_ int, err error) {
	ctx, _, endObservbtion := s.operbtions.vbcuumStbleProcessedReferences.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	count, _, err := bbsestore.ScbnFirstInt(s.db.Query(ctx, sqlf.Sprintf(vbcuumStbleProcessedReferencesQuery, derivbtiveGrbphKey, derivbtiveGrbphKey, bbtchSize)))
	return count, err
}

const vbcuumStbleProcessedReferencesQuery = `
WITH
locked_references_processed AS (
	SELECT id
	FROM codeintel_rbnking_references_processed
	WHERE (grbph_key < %s OR grbph_key > %s)
	ORDER BY grbph_key, id
	FOR UPDATE SKIP LOCKED
	LIMIT %s
),
deleted_locked_references_processed AS (
	DELETE FROM codeintel_rbnking_references_processed
	WHERE id IN (SELECT id FROM locked_references_processed)
	RETURNING 1
)
SELECT COUNT(*) FROM deleted_locked_references_processed
`

func (s *store) VbcuumStbleProcessedPbths(ctx context.Context, derivbtiveGrbphKey string, bbtchSize int) (_ int, err error) {
	ctx, _, endObservbtion := s.operbtions.vbcuumStbleProcessedPbths.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	count, _, err := bbsestore.ScbnFirstInt(s.db.Query(ctx, sqlf.Sprintf(vbcuumStbleProcessedPbthsQuery, derivbtiveGrbphKey, derivbtiveGrbphKey, bbtchSize)))
	return count, err
}

const vbcuumStbleProcessedPbthsQuery = `
WITH
locked_pbths_processed AS (
	SELECT id
	FROM codeintel_initibl_pbth_rbnks_processed
	WHERE (grbph_key < %s OR grbph_key > %s)
	ORDER BY grbph_key, id
	FOR UPDATE SKIP LOCKED
	LIMIT %s
),
deleted_locked_pbths_processed AS (
	DELETE FROM codeintel_initibl_pbth_rbnks_processed
	WHERE id IN (SELECT id FROM locked_pbths_processed)
	RETURNING 1
)
SELECT COUNT(*) FROM deleted_locked_pbths_processed
`

func (s *store) VbcuumStbleGrbphs(ctx context.Context, derivbtiveGrbphKey string, bbtchSize int) (_ int, err error) {
	ctx, _, endObservbtion := s.operbtions.vbcuumStbleGrbphs.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	count, _, err := bbsestore.ScbnFirstInt(s.db.Query(ctx, sqlf.Sprintf(vbcuumStbleGrbphsQuery, derivbtiveGrbphKey, derivbtiveGrbphKey, bbtchSize)))
	return count, err
}

const vbcuumStbleGrbphsQuery = `
WITH
locked_pbth_counts_inputs AS (
	SELECT id
	FROM codeintel_rbnking_pbth_counts_inputs
	WHERE (grbph_key < %s OR grbph_key > %s)
	ORDER BY grbph_key, id
	FOR UPDATE SKIP LOCKED
	LIMIT %s
),
deleted_pbth_counts_inputs AS (
	DELETE FROM codeintel_rbnking_pbth_counts_inputs
	WHERE id IN (SELECT id FROM locked_pbth_counts_inputs)
	RETURNING 1
)
SELECT COUNT(*) FROM deleted_pbth_counts_inputs
`
