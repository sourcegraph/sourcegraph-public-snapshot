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
func CreateTemporaryTable(ctx context.Context, db dbutil.DB, tableName string, selectFields []ColumnAndType) error {
	query := CreateTemporaryTableQuery(tableName, selectFields)
	_, err := db.ExecContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	return err
}

// TODO - test
// TODO - document
func CreateTemporaryTableQuery(tableName string, selectFields []ColumnAndType) *sqlf.Query {
	namesAndTypes := make([]*sqlf.Query, 0, len(selectFields))
	for _, field := range selectFields {
		namesAndTypes = append(namesAndTypes, sqlf.Sprintf(field.Name+" "+field.PostgresType))
	}

	return sqlf.Sprintf(
		"CREATE TEMPORARY TABLE %s (%s) ON COMMIT DROP",
		sqlf.Sprintf(tableName),
		sqlf.Join(namesAndTypes, ", "),
	)
}

// TODO - test
// TODO - document
func UpdateFromTemporaryTable(
	ctx context.Context,
	db dbutil.DB,
	sourceTableName string,
	destinationTableName string,
	primaryKeyFields []string,
	constantPrimaryKeyValues map[string]interface{},
	assignmentFields []string,
	constantAssignmentValues map[string]interface{},
) error {
	query := UpdateFromTemporaryTableQuery(
		sourceTableName,
		destinationTableName,
		primaryKeyFields,
		constantPrimaryKeyValues,
		assignmentFields,
		constantAssignmentValues,
	)

	_, err := db.ExecContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	return err
}

// TODO - test
// TODO - document
func UpdateFromTemporaryTableQuery(
	sourceTableName string,
	destinationTableName string,
	primaryKeyFields []string, // where dest.X = src.X
	constantPrimaryKeyValues map[string]interface{}, // where dest.X = value
	assignmentFields []string, // (X = src.X) And (where X != src.X)
	constantAssignmentValues map[string]interface{}, // (X = value) and (where X != value)
) *sqlf.Query {
	assignments := make([]*sqlf.Query, 0, len(assignmentFields)+len(constantAssignmentValues))
	conditions := make([]*sqlf.Query, 0, len(primaryKeyFields)+len(constantPrimaryKeyValues))
	assignmentConditions := make([]*sqlf.Query, 0, len(assignmentFields)+len(constantAssignmentValues))

	for _, name := range primaryKeyFields {
		conditions = append(conditions, sqlf.Sprintf("dest."+name+"= src."+name))
	}
	for k, v := range constantPrimaryKeyValues {
		conditions = append(conditions, sqlf.Sprintf("dest."+k+" = %s", v))
	}
	for _, name := range assignmentFields {
		assignments = append(assignments, sqlf.Sprintf(name+" = src."+name))
		assignmentConditions = append(assignmentConditions, sqlf.Sprintf("dest."+name+" = src."+name))
	}
	for k, v := range constantAssignmentValues {
		query := sqlf.Sprintf(k+" = %s", v)
		assignments = append(assignments, query)
		assignmentConditions = append(assignmentConditions, query)
	}

	return sqlf.Sprintf(
		"UPDATE %s dest SET %s FROM %s src WHERE %s AND NOT (%s)",
		sqlf.Sprintf(destinationTableName),
		sqlf.Join(assignments, ", "),
		sqlf.Sprintf(sourceTableName),
		sqlf.Join(conditions, " AND "),
		sqlf.Join(assignmentConditions, " AND "),
	)
}
