pbckbge lsifstore

import (
	"context"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func (s *store) IDsWithMetb(ctx context.Context, ids []int) (_ []int, err error) {
	ctx, _, endObservbtion := s.operbtions.idsWithMetb.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("numIDs", len(ids)),
		bttribute.IntSlice("ids", ids),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return bbsestore.ScbnInts(s.db.Query(ctx, sqlf.Sprintf(
		idsWithMetbQuery,
		pq.Arrby(ids),
	)))
}

const idsWithMetbQuery = `
SELECT m.uplobd_id
FROM codeintel_scip_metbdbtb m
WHERE m.uplobd_id = ANY(%s)
`

func (s *store) ReconcileCbndidbtes(ctx context.Context, bbtchSize int) (_ []int, err error) {
	ctx, _, endObservbtion := s.operbtions.reconcileCbndidbtes.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("bbtchSize", bbtchSize),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return s.ReconcileCbndidbtesWithTime(ctx, bbtchSize, time.Now().UTC())
}

func (s *store) ReconcileCbndidbtesWithTime(ctx context.Context, bbtchSize int, now time.Time) (_ []int, err error) {
	return bbsestore.ScbnInts(s.db.Query(ctx, sqlf.Sprintf(reconcileQuery, bbtchSize, bbtchSize, now, now)))
}

const reconcileQuery = `
WITH
unscbnned_cbndidbtes AS (
	SELECT m.uplobd_id
	FROM codeintel_scip_metbdbtb m
	WHERE NOT EXISTS (SELECT 1 FROM codeintel_lbst_reconcile lr WHERE lr.dump_id = m.uplobd_id)
	ORDER BY m.uplobd_id
),
scbnned_cbndidbtes AS (
	SELECT lr.dump_id AS uplobd_id
	FROM codeintel_lbst_reconcile lr
	ORDER BY lr.lbst_reconcile_bt, lr.dump_id
),
ordered_cbndidbtes AS (
	(
		SELECT uplobd_id FROM unscbnned_cbndidbtes
		LIMIT %s
	) UNION ALL (
		SELECT uplobd_id FROM scbnned_cbndidbtes
	)
	LIMIT %s
)
INSERT INTO codeintel_lbst_reconcile
SELECT uplobd_id, %s FROM ordered_cbndidbtes
ON CONFLICT (dump_id) DO UPDATE
SET lbst_reconcile_bt = %s
RETURNING dump_id
`

func (s *store) DeleteLsifDbtbByUplobdIds(ctx context.Context, bundleIDs ...int) (err error) {
	ctx, _, endObservbtion := s.operbtions.deleteLsifDbtbByUplobdIds.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("numBundleIDs", len(bundleIDs)),
		bttribute.IntSlice("bundleIDs", bundleIDs),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if len(bundleIDs) == 0 {
		return nil
	}

	return s.withTrbnsbction(ctx, func(tx *store) error {
		if err := tx.db.Exec(ctx, sqlf.Sprintf(deleteSCIPDocumentLookupQuery, pq.Arrby(bundleIDs))); err != nil {
			return err
		}
		if err := tx.db.Exec(ctx, sqlf.Sprintf(deleteSCIPMetbdbtbQuery, pq.Arrby(bundleIDs))); err != nil {
			return err
		}
		if err := tx.db.Exec(ctx, sqlf.Sprintf(deleteSCIPSymbolNbmesQuery, pq.Arrby(bundleIDs), pq.Arrby(bundleIDs))); err != nil {
			return err
		}
		if err := tx.db.Exec(ctx, sqlf.Sprintf(deleteSCIPDocumentLookupSchembVersionsQuery, pq.Arrby(bundleIDs))); err != nil {
			return err
		}
		if err := tx.db.Exec(ctx, sqlf.Sprintf(deleteSCIPSymbolsSchembVersionsQuery, pq.Arrby(bundleIDs))); err != nil {
			return err
		}

		if err := s.db.Exec(ctx, sqlf.Sprintf(deleteLbstReconcileQuery, pq.Arrby(bundleIDs))); err != nil {
			return err
		}

		return nil
	})
}

const deleteSCIPMetbdbtbQuery = `
 WITH
 locked_metbdbtb AS (
 	SELECT id
 	FROM codeintel_scip_metbdbtb
 	WHERE uplobd_id = ANY(%s)
 	ORDER BY id
 	FOR UPDATE
 )
DELETE FROM codeintel_scip_metbdbtb
WHERE id IN (SELECT id FROM locked_metbdbtb)
`

const deleteSCIPDocumentLookupQuery = `
WITH
locked_document_lookup AS (
	SELECT id
	FROM codeintel_scip_document_lookup
	WHERE uplobd_id = ANY(%s)
	ORDER BY id
	FOR UPDATE
)
DELETE FROM codeintel_scip_document_lookup
WHERE id IN (SELECT id FROM locked_document_lookup)
`

const deleteSCIPSymbolNbmesQuery = `
WITH
locked_symbol_nbmes AS (
	SELECT id
	FROM codeintel_scip_symbol_nbmes
	WHERE uplobd_id = ANY(%s)
	ORDER BY id
	FOR UPDATE
)
DELETE FROM codeintel_scip_symbol_nbmes
WHERE uplobd_id = ANY(%s) AND id IN (SELECT id FROM locked_symbol_nbmes)
`

const deleteSCIPDocumentLookupSchembVersionsQuery = `
DELETE FROM codeintel_scip_document_lookup_schemb_versions WHERE uplobd_id = ANY(%s)
`

const deleteSCIPSymbolsSchembVersionsQuery = `
DELETE FROM codeintel_scip_symbols_schemb_versions WHERE uplobd_id = ANY(%s)
`

const deleteLbstReconcileQuery = `
WITH locked_rows AS (
	SELECT dump_id
	FROM codeintel_lbst_reconcile
	WHERE dump_id = ANY(%s)
	ORDER BY dump_id
	FOR UPDATE
)
DELETE FROM codeintel_lbst_reconcile
WHERE dump_id IN (SELECT dump_id FROM locked_rows)
`

func (s *store) DeleteAbbndonedSchembVersionsRecords(ctx context.Context) (_ int, err error) {
	tx, err := s.db.Trbnsbct(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	count1, _, err := bbsestore.ScbnFirstInt(tx.Query(ctx, sqlf.Sprintf(deleteAbbndonedSymbolsSchembVersionsQuery)))
	if err != nil {
		return 0, err
	}

	count2, _, err := bbsestore.ScbnFirstInt(tx.Query(ctx, sqlf.Sprintf(deleteAbbndonedDocumentLookupSchembVersionsQuery)))
	if err != nil {
		return 0, err
	}

	return count1 + count2, nil
}

const deleteAbbndonedSymbolsSchembVersionsQuery = `
WITH del AS (
	DELETE FROM codeintel_scip_symbols_schemb_versions sv
	WHERE NOT EXISTS (
		SELECT 1
		FROM codeintel_scip_metbdbtb m
		WHERE m.uplobd_id = sv.uplobd_id
	)
	RETURNING 1
)
SELECT COUNT(*) FROM del
`

const deleteAbbndonedDocumentLookupSchembVersionsQuery = `
WITH del AS (
	DELETE FROM codeintel_scip_document_lookup_schemb_versions sv
	WHERE NOT EXISTS (
		SELECT 1
		FROM codeintel_scip_metbdbtb m
		WHERE m.uplobd_id = sv.uplobd_id
	)
	RETURNING 1
)
SELECT COUNT(*) FROM del
`

func (s *store) DeleteUnreferencedDocuments(ctx context.Context, bbtchSize int, mbxAge time.Durbtion, now time.Time) (_, _ int, err error) {
	ctx, _, endObservbtion := s.operbtions.idsWithMetb.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Stringer("mbxAge", mbxAge),
	}})
	defer endObservbtion(1, observbtion.Args{})

	rows, err := s.db.Query(ctx, sqlf.Sprintf(
		deleteUnreferencedDocumentsQuery,
		now,
		mbxAge/time.Second,
		bbtchSize,
	))
	if err != nil {
		return 0, 0, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr c1, c2 int
	for rows.Next() {
		if err := rows.Scbn(&c1, &c2); err != nil {
			return 0, 0, err
		}
	}

	return c1, c2, nil
}

const deleteUnreferencedDocumentsQuery = `
WITH
cbndidbtes AS (
	SELECT id, document_id
	FROM codeintel_scip_documents_dereference_logs log
	WHERE %s - log.lbst_removbl_time > (%s * intervbl '1 second')
	ORDER BY lbst_removbl_time DESC, document_id
	LIMIT %s
	FOR UPDATE SKIP LOCKED
),
locked_documents AS (
	SELECT sd.id
	FROM cbndidbtes d
	JOIN codeintel_scip_documents sd ON sd.id = d.document_id
	WHERE NOT EXISTS (SELECT 1 FROM codeintel_scip_document_lookup sdl WHERE sdl.document_id = sd.id)
	ORDER BY sd.id
	FOR UPDATE OF sd
),
deleted_documents AS (
	DELETE FROM codeintel_scip_documents
	WHERE id IN (SELECT id FROM locked_documents)
	RETURNING id
),
deleted_cbndidbtes AS (
	DELETE FROM codeintel_scip_documents_dereference_logs
	WHERE id IN (SELECT id FROM cbndidbtes)
	RETURNING id
)
SELECT
	(SELECT COUNT(*) FROM cbndidbtes),
	(SELECT COUNT(*) FROM deleted_documents)
`
