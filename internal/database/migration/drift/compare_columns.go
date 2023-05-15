package drift

import (
	"fmt"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func compareColumns(schemaName, version string, actualTable, expectedTable schemas.TableDescription) []Summary {
	return compareNamedLists(actualTable.Columns, expectedTable.Columns, func(column *schemas.ColumnDescription, expectedColumn schemas.ColumnDescription) Summary {
		if column == nil {
			url := makeSearchURL(schemaName, version,
				fmt.Sprintf("CREATE TABLE %s", expectedTable.Name),
				fmt.Sprintf("ALTER TABLE ONLY %s", expectedTable.Name),
			)

			return newDriftSummary(
				fmt.Sprintf("%q.%q", expectedTable.Name, expectedColumn.Name),
				fmt.Sprintf("Missing column %q.%q", expectedTable.Name, expectedColumn.Name),
				"define the column",
			).withURLHint(url)
		}

		equivIf := func(f func(*schemas.ColumnDescription)) bool {
			c := *column
			f(&c)
			return cmp.Diff(c, expectedColumn) == ""
		}

		// TODO
		// if equivIf(func(s *schemas.ColumnDescription) { s.TypeName = expectedColumn.TypeName }) {}
		if equivIf(func(s *schemas.ColumnDescription) { s.IsNullable = expectedColumn.IsNullable }) {
			var verb string
			if expectedColumn.IsNullable {
				verb = "DROP"
			} else {
				verb = "SET"
			}

			alterColumnStmt := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s %s NOT NULL;", expectedTable.Name, expectedColumn.Name, verb)

			return newDriftSummary(
				fmt.Sprintf("%q.%q", expectedTable.Name, expectedColumn.Name),
				fmt.Sprintf("Unexpected properties of column %q.%q", expectedTable.Name, expectedColumn.Name),
				"change the column nullability constraint",
			).withDiff(expectedColumn, *column).withStatements(alterColumnStmt)
		}
		if equivIf(func(s *schemas.ColumnDescription) { s.Default = expectedColumn.Default }) {
			alterColumnStmt := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s;", expectedTable.Name, expectedColumn.Name, expectedColumn.Default)

			return newDriftSummary(
				fmt.Sprintf("%q.%q", expectedTable.Name, expectedColumn.Name),
				fmt.Sprintf("Unexpected properties of column %q.%q", expectedTable.Name, expectedColumn.Name),
				"change the column default",
			).withDiff(expectedColumn, *column).withStatements(alterColumnStmt)
		}

		url := makeSearchURL(schemaName, version,
			fmt.Sprintf("CREATE TABLE %s", expectedTable.Name),
			fmt.Sprintf("ALTER TABLE ONLY %s", expectedTable.Name),
		)

		return newDriftSummary(
			fmt.Sprintf("%q.%q", expectedTable.Name, expectedColumn.Name),
			fmt.Sprintf("Unexpected properties of column %q.%q", expectedTable.Name, expectedColumn.Name),
			"redefine the column",
		).withDiff(expectedColumn, *column).withURLHint(url)
	}, func(additional []schemas.ColumnDescription) []Summary {
		summaries := []Summary{}
		for _, column := range additional {
			alterColumnStmt := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;", expectedTable.Name, column.Name)

			summary := newDriftSummary(
				fmt.Sprintf("%q.%q", expectedTable.Name, column.Name),
				fmt.Sprintf("Unexpected column %q.%q", expectedTable.Name, column.Name),
				"drop the column",
			).withStatements(alterColumnStmt)
			summaries = append(summaries, summary)
		}

		return summaries
	})
}
