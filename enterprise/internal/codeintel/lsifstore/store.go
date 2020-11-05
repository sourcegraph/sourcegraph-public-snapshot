package lsifstore

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
)

type Store interface {
	Clear(ctx context.Context, bundleIDs ...int) error
}

type store struct {
	*basestore.Store
}

func New(db dbutil.DB) Store {
	return &store{
		Store: basestore.NewWithHandle(basestore.NewHandleWithDB(db, sql.TxOptions{})),
	}
}

var tableNames = []string{
	"lsif_data_metadata",
	"lsif_data_documents",
	"lsif_data_result_chunks",
	"lsif_data_definitions",
	"lsif_data_references",
}

func (s *store) Clear(ctx context.Context, bundleIDs ...int) (err error) {
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
