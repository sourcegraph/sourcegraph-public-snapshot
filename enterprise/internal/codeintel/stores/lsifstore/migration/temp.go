package migration

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
)

type ColumnAndType struct {
	// Name is the name of the column.
	Name string

	// PostgresType is the type (and modifiers) of the column.
	PostgresType string
}

// TODO - extract
// TODO - document
// TODO - break down even more
func BulkInsertIntoTemporaryTable(
	ctx context.Context,
	store *basestore.Store,
	selectFields []ColumnAndType,
	values <-chan []interface{},
) error {
	names := make([]string, 0, len(selectFields))
	namesAndTypes := make([]*sqlf.Query, 0, len(selectFields))

	for _, field := range selectFields {
		names = append(names, field.Name)
		namesAndTypes = append(namesAndTypes, sqlf.Sprintf(field.Name+" "+field.PostgresType))
	}

	if err := store.Exec(ctx, sqlf.Sprintf(bulkInsertIntoTemporaryTableQuery, sqlf.Join(namesAndTypes, ", "))); err != nil {
		return err
	}

	inserter := batch.NewBatchInserter(ctx, store.Handle().DB(), "t_target", names...)

	for row := range values {
		if err := inserter.Insert(ctx, row...); err != nil {
			return err
		}
	}
	if err := inserter.Flush(ctx); err != nil {
		return err
	}

	return nil
}

const bulkInsertIntoTemporaryTableQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/migration/temp.go:BulkInsertIntoTemporaryTable
CREATE TEMPORARY TABLE t_target (%s) ON COMMIT DROP
`

// TODO - extract
// TODO - document
func BulkUpdateFromTempTable(
	ctx context.Context,
	store *basestore.Store,
	tableName string,
	primaryKeyFields []string,
	updateFields []string,
	constantFieldValues map[string]interface{},
) error {
	conditions := make([]*sqlf.Query, 0, len(primaryKeyFields))
	for _, name := range primaryKeyFields {
		// disambiguate fields
		conditions = append(conditions, sqlf.Sprintf("dest."+name+"= src."+name))
	}

	assignments := make([]*sqlf.Query, 0, len(updateFields))
	for _, name := range updateFields {
		// TODO
		assignments = append(assignments, sqlf.Sprintf(name+"= src."+name))
	}

	for k, v := range constantFieldValues {
		// TODO
		assignments = append(assignments, sqlf.Sprintf("%s = %s", sqlf.Sprintf(k), v))
	}

	return store.Exec(ctx, sqlf.Sprintf(
		bulkUpdateFromTempTableQuery,
		sqlf.Sprintf(tableName),
		sqlf.Join(assignments, ", "),
		sqlf.Join(conditions, " AND "),
	))
}

const bulkUpdateFromTempTableQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/migration/temp.go:BulkUpdateFromTempTable
UPDATE %s dest SET %s FROM t_target src WHERE %s
`
