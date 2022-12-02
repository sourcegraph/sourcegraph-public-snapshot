package lsifstore

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// DeleteLsifDataByUploadIds deletes LSIF data by UploadIds from the lsif database.
func (s *store) DeleteLsifDataByUploadIds(ctx context.Context, bundleIDs ...int) (err error) {
	ctx, _, endObservation := s.operations.deleteLsifDataByUploadIds.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Int("numBundleIDs", len(bundleIDs)),
		otlog.String("bundleIDs", intsToString(bundleIDs)),
	}})
	defer endObservation(1, observation.Args{})

	if len(bundleIDs) == 0 {
		return nil
	}

	if err := s.deleteLSIFData(ctx, bundleIDs); err != nil {
		return err
	}

	if err := s.deleteSCIPData(ctx, bundleIDs); err != nil {
		return err
	}

	if err := s.db.Exec(ctx, sqlf.Sprintf(deleteLastReconcileQuery, pq.Array(bundleIDs))); err != nil {
		return err
	}

	return nil
}

const deleteLastReconcileQuery = `
WITH locked_rows AS (
	SELECT dump_id
	FROM codeintel_last_reconcile
	WHERE dump_id = ANY(%s)
	ORDER BY dump_id
	FOR UPDATE
)
DELETE FROM codeintel_last_reconcile WHERE dump_id IN (SELECT dump_id FROM locked_rows)
`

var lsifDataTables = []string{
	"lsif_data_metadata",
	"lsif_data_documents",
	"lsif_data_result_chunks",
	"lsif_data_definitions",
	"lsif_data_references",
	"lsif_data_implementations",
}

func (s *store) deleteLSIFData(ctx context.Context, uploadIDs []int) error {
	tx, err := s.db.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Done(err)
	}()

	// Ensure ids are sorted so that we take row locks during the DELETE query
	// in a deterministic order. This should prevent deadlocks with other queries
	// that mass update the same table.
	sort.Ints(uploadIDs)

	for _, tableName := range lsifDataTables {
		query := sqlf.Sprintf(`DELETE FROM %s WHERE dump_id = ANY(%s)`, sqlf.Sprintf(tableName), pq.Array(uploadIDs))
		if err := tx.Exec(ctx, query); err != nil {
			return err
		}

	}

	return nil
}

func (s *store) deleteSCIPData(ctx context.Context, uploadIDs []int) error {
	tx, err := s.db.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Done(err)
	}()

	if err := tx.Exec(ctx, sqlf.Sprintf(deleteSCIPDocumentLookupQuery, pq.Array(uploadIDs))); err != nil {
		return err
	}
	if err := tx.Exec(ctx, sqlf.Sprintf(deleteSCIPMetadataQuery, pq.Array(uploadIDs))); err != nil {
		return err
	}

	return nil
}

const deleteSCIPMetadataQuery = `
 WITH
 locked_metadata AS (
 	SELECT id
 	FROM codeintel_scip_metadata
 	WHERE upload_id = ANY(%s)
 	ORDER BY id
 	FOR UPDATE
 )
DELETE FROM codeintel_scip_metadata
WHERE id IN (SELECT id FROM locked_metadata)
`

const deleteSCIPDocumentLookupQuery = `
WITH
locked_document_lookup AS (
	SELECT id
	FROM codeintel_scip_document_lookup
	WHERE upload_id = ANY(%s)
	ORDER BY id
	FOR UPDATE
)
DELETE FROM codeintel_scip_document_lookup
WHERE id IN (SELECT id FROM locked_document_lookup)
`

func intsToString(vs []int) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, strconv.Itoa(v))
	}

	return strings.Join(strs, ", ")
}
