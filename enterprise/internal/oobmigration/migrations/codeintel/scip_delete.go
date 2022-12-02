package codeintel

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

var lsifTableNames = []string{
	"lsif_data_metadata",
	"lsif_data_documents",
	"lsif_data_result_chunks",
	"lsif_data_definitions",
	"lsif_data_references",
	"lsif_data_implementations",
}

func deleteLSIFData(ctx context.Context, tx *basestore.Store, uploadID int) error {
	for _, tableName := range lsifTableNames {
		if err := tx.Exec(ctx, sqlf.Sprintf(
			deleteLSIFDataQuery,
			sqlf.Sprintf(tableName),
			uploadID,
		)); err != nil {
			return err
		}
	}

	return nil
}

const deleteLSIFDataQuery = `
DELETE FROM %s WHERE dump_id = %s
`
