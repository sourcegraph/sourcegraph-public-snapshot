package lsifstore

import (
	"context"

	"github.com/keegancsmith/sqlf"
)

var tableNames = []string{
	"lsif_data_metadata",
	"lsif_data_documents",
	"lsif_data_result_chunks",
	"lsif_data_definitions",
	"lsif_data_references",
}

func (s *Store) Clear(ctx context.Context, bundleIDs ...int) (err error) {
	if len(bundleIDs) == 0 {
		return nil
	}

	var ids []*sqlf.Query
	for _, bundleID := range bundleIDs {
		ids = append(ids, sqlf.Sprintf("%d", bundleID))
	}

	tx, err := s.Store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Done(err)
	}()

	for _, tableName := range tableNames {
		if err := tx.Exec(ctx, sqlf.Sprintf(`DELETE FROM "`+tableName+`" WHERE dump_id IN (%s)`, sqlf.Join(ids, ","))); err != nil {
			return err
		}
	}

	return nil
}
