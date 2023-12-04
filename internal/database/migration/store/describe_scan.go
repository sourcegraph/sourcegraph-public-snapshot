package store

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type Extension struct {
	SchemaName    string
	ExtensionName string
}

func scanExtensions(rows *sql.Rows, queryErr error) (_ []Extension, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var extensions []Extension
	for rows.Next() {
		var extension Extension
		if err := rows.Scan(&extension.SchemaName, &extension.ExtensionName); err != nil {
			return nil, err
		}

		extensions = append(extensions, extension)
	}

	return extensions, nil
}

type enum struct {
	SchemaName string
	TypeName   string
	Label      string
}

func scanEnums(rows *sql.Rows, queryErr error) (_ []enum, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var enums []enum
	for rows.Next() {
		var enum enum

		if err := rows.Scan(
			&enum.SchemaName,
			&enum.TypeName,
			&enum.Label,
		); err != nil {
			return nil, err
		}

		enums = append(enums, enum)
	}

	return enums, nil
}

type function struct {
	SchemaName   string
	FunctionName string
	Fancy        string
	ReturnType   string
	Definition   string
}

func scanFunctions(rows *sql.Rows, queryErr error) (_ []function, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var functions []function
	for rows.Next() {
		var function function

		if err := rows.Scan(
			&function.SchemaName,
			&function.FunctionName,
			&function.Fancy,
			&function.ReturnType,
			&function.Definition,
		); err != nil {
			return nil, err
		}

		functions = append(functions, function)
	}

	return functions, nil
}

type sequence struct {
	SchemaName   string
	SequenceName string
	DataType     string
	StartValue   int
	MinimumValue int
	MaximumValue int
	Increment    int
	CycleOption  string
}

func scanSequences(rows *sql.Rows, queryErr error) (_ []sequence, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var sequences []sequence
	for rows.Next() {
		var sequence sequence

		if err := rows.Scan(
			&sequence.SchemaName,
			&sequence.SequenceName,
			&sequence.DataType,
			&sequence.StartValue,
			&sequence.MinimumValue,
			&sequence.MaximumValue,
			&sequence.Increment,
			&sequence.CycleOption,
		); err != nil {
			return nil, err
		}

		sequences = append(sequences, sequence)
	}

	return sequences, nil
}

type table struct {
	SchemaName string
	TableName  string
	Comment    string
}

func scanTables(rows *sql.Rows, queryErr error) (_ []table, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var tables []table
	for rows.Next() {
		var table table
		if err := rows.Scan(
			&table.SchemaName,
			&table.TableName,
			&dbutil.NullString{S: &table.Comment},
		); err != nil {
			return nil, err
		}

		tables = append(tables, table)
	}

	return tables, nil
}

type column struct {
	SchemaName             string
	TableName              string
	ColumnName             string
	Index                  int
	DataType               string
	IsNullable             bool
	Default                string
	CharacterMaximumLength int
	IsIdentity             bool
	IdentityGeneration     string
	IsGenerated            string
	GenerationExpression   string
	Comment                string
}

func scanColumns(rows *sql.Rows, queryErr error) (_ []column, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var columns []column
	for rows.Next() {
		var (
			column     column
			isNullable string
			isIdentity string
		)

		if err := rows.Scan(
			&column.SchemaName,
			&column.TableName,
			&column.ColumnName,
			&column.Index,
			&column.DataType,
			&isNullable,
			&dbutil.NullString{S: &column.Default},
			&dbutil.NullInt{N: &column.CharacterMaximumLength},
			&isIdentity,
			&dbutil.NullString{S: &column.IdentityGeneration},
			&column.IsGenerated,
			&dbutil.NullString{S: &column.GenerationExpression},
			&dbutil.NullString{S: &column.Comment},
		); err != nil {
			return nil, err
		}

		column.IsNullable = isTruthy(isNullable)
		column.IsIdentity = isTruthy(isIdentity)
		columns = append(columns, column)
	}

	return columns, nil
}

type index struct {
	SchemaName           string
	TableName            string
	IndexName            string
	IsPrimaryKey         bool
	IsUnique             bool
	IsExclusion          bool
	IsDeferrable         bool
	IndexDefinition      string
	ConstraintType       string
	ConstraintDefinition string
}

func scanIndexes(rows *sql.Rows, queryErr error) (_ []index, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var indexes []index
	for rows.Next() {
		var (
			index        index
			isPrimaryKey string
			isUnique     string
		)

		if err := rows.Scan(
			&index.SchemaName,
			&index.TableName,
			&index.IndexName,
			&isPrimaryKey,
			&isUnique,
			&dbutil.NullBool{B: &index.IsExclusion},
			&dbutil.NullBool{B: &index.IsDeferrable},
			&index.IndexDefinition,
			&dbutil.NullString{S: &index.ConstraintType},
			&dbutil.NullString{S: &index.ConstraintDefinition},
		); err != nil {
			return nil, err
		}

		index.IsPrimaryKey = isTruthy(isPrimaryKey)
		index.IsUnique = isTruthy(isUnique)
		indexes = append(indexes, index)
	}

	return indexes, nil
}

type constraint struct {
	SchemaName           string
	TableName            string
	ConstraintName       string
	ConstraintType       string
	IsDeferrable         bool
	RefTableName         string
	ConstraintDefinition string
}

func scanConstraints(rows *sql.Rows, queryErr error) (_ []constraint, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var constraints []constraint
	for rows.Next() {
		var constraint constraint

		if err := rows.Scan(
			&constraint.SchemaName,
			&constraint.TableName,
			&constraint.ConstraintName,
			&constraint.ConstraintType,
			&dbutil.NullBool{B: &constraint.IsDeferrable},
			&dbutil.NullString{S: &constraint.RefTableName},
			&constraint.ConstraintDefinition,
		); err != nil {
			return nil, err
		}

		constraints = append(constraints, constraint)
	}

	return constraints, nil
}

type trigger struct {
	SchemaName        string
	TableName         string
	TriggerName       string
	TriggerDefinition string
}

func scanTriggers(rows *sql.Rows, queryErr error) (_ []trigger, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var triggers []trigger
	for rows.Next() {
		var trigger trigger

		if err := rows.Scan(
			&trigger.SchemaName,
			&trigger.TableName,
			&trigger.TriggerName,
			&trigger.TriggerDefinition,
		); err != nil {
			return nil, err
		}

		triggers = append(triggers, trigger)
	}

	return triggers, nil
}

type view struct {
	SchemaName string
	ViewName   string
	Definition string
}

func scanViews(rows *sql.Rows, queryErr error) (_ []view, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var views []view
	for rows.Next() {
		var view view

		if err := rows.Scan(
			&view.SchemaName,
			&view.ViewName,
			&view.Definition,
		); err != nil {
			return nil, err
		}

		views = append(views, view)
	}

	return views, nil
}
