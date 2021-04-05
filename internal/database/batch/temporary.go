package batch

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type ColumnAndType struct {
	// Name is the name of the column.
	Name string

	// PostgresType is the type (and modifiers) of the column.
	PostgresType string
}

// TODO - test
// TODO - document
func CreateTemporaryTable(
	ctx context.Context,
	db dbutil.DB,
	tableName string,
	selectFields []ColumnAndType,
) error {
	namesAndTypes := make([]*sqlf.Query, 0, len(selectFields))
	for _, field := range selectFields {
		namesAndTypes = append(namesAndTypes, sqlf.Sprintf(field.Name+" "+field.PostgresType))
	}

	query := sqlf.Sprintf(
		"CREATE TEMPORARY TABLE %s (%s) ON COMMIT DROP",
		sqlf.Sprintf(tableName),
		sqlf.Join(namesAndTypes, ", "),
	)
	_, err := db.ExecContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	return err
}

// TODO - test
// TODO - document
func UpdateFromTemporaryTable(
	ctx context.Context,
	db dbutil.DB,
	sourceTableName string,
	destinationTableName string,
	primaryKeyFields []string,
	updateFields []string,
	constantFieldValues map[string]interface{},
) error {
	assignments := make([]*sqlf.Query, 0, len(updateFields)+len(constantFieldValues))
	for _, name := range updateFields {
		assignments = append(assignments, sqlf.Sprintf(name+" = src."+name))
	}
	for k, v := range constantFieldValues {
		assignments = append(assignments, sqlf.Sprintf(k+" = %s", v))
	}

	conditions := make([]*sqlf.Query, 0, len(primaryKeyFields))
	for _, name := range primaryKeyFields {
		// use prefix to disambiguate fields
		conditions = append(conditions, sqlf.Sprintf("dest."+name+"= src."+name))
	}

	query := sqlf.Sprintf(
		"UPDATE %s dest SET %s FROM %s src WHERE %s",
		sqlf.Sprintf(destinationTableName),
		sqlf.Join(assignments, ", "),
		sqlf.Sprintf(sourceTableName),
		sqlf.Join(conditions, " AND "),
	)
	_, err := db.ExecContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	return err
}
