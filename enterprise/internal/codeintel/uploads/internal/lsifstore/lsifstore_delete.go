package lsifstore

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/keegancsmith/sqlf"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

var tableNames = map[string]string{
	"lsif_data_metadata":                        "dump_id",
	"lsif_data_documents":                       "dump_id",
	"lsif_data_documents_schema_versions":       "dump_id",
	"lsif_data_result_chunks":                   "dump_id",
	"lsif_data_definitions":                     "dump_id",
	"lsif_data_definitions_schema_versions":     "dump_id",
	"lsif_data_references":                      "dump_id",
	"lsif_data_references_schema_versions":      "dump_id",
	"lsif_data_implementations":                 "dump_id",
	"lsif_data_implementations_schema_versions": "dump_id",
	"codeintel_last_reconcile":                  "dump_id",
	"codeintel_scip_metadata":                   "upload_id",
	"codeintel_scip_document_lookup":            "upload_id",
	"codeintel_scip_symbols":                    "upload_id",
}

// DeleteLsifDataByUploadIds deletes LSIF data by UploadIds from the lsif database.
func (s *store) DeleteLsifDataByUploadIds(ctx context.Context, bundleIDs ...int) (err error) {
	ctx, trace, endObservation := s.operations.deleteLsifDataByUploadIds.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.Int("numBundleIDs", len(bundleIDs)),
		otlog.String("bundleIDs", intsToString(bundleIDs)),
	}})
	defer endObservation(1, observation.Args{})

	if len(bundleIDs) == 0 {
		return nil
	}

	// Ensure ids are sorted so that we take row locks during the
	// DELETE query in a determinstic order. This should prevent
	// deadlocks with other queries that mass update the same table.
	sort.Ints(bundleIDs)

	var ids []*sqlf.Query
	for _, bundleID := range bundleIDs {
		ids = append(ids, sqlf.Sprintf("%d", bundleID))
	}

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Done(err)
	}()

	for tableName, fieldName := range tableNames {
		trace.Log(otlog.String("tableName", tableName))

		query := sqlf.Sprintf(deleteQuery, sqlf.Sprintf(tableName), sqlf.Sprintf(fieldName), sqlf.Join(ids, ","))
		if err := tx.Exec(ctx, query); err != nil {
			return err
		}
	}

	return nil
}

const deleteQuery = `
DELETE FROM %s WHERE %s IN (%s)
`

func intsToString(vs []int) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, strconv.Itoa(v))
	}

	return strings.Join(strs, ", ")
}
