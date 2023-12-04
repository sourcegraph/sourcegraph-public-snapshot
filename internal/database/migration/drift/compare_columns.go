package drift

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func compareColumns(schemaName, version string, actualTable, expectedTable schemas.TableDescription) []Summary {
	return compareNamedListsStrict(
		actualTable.Columns,
		expectedTable.Columns,
		compareColumnsCallbackFor(schemaName, version, expectedTable),
		compareColumnsAdditionalCallbackFor(expectedTable),
	)
}

func compareColumnsCallbackFor(schemaName, version string, table schemas.TableDescription) func(_ *schemas.ColumnDescription, _ schemas.ColumnDescription) Summary {
	return func(column *schemas.ColumnDescription, expectedColumn schemas.ColumnDescription) Summary {
		if column == nil {
			return newDriftSummary(
				fmt.Sprintf("%q.%q", table.GetName(), expectedColumn.GetName()),
				fmt.Sprintf("Missing column %q.%q", table.GetName(), expectedColumn.GetName()),
				"define the column",
			).withStatements(
				expectedColumn.CreateStatement(table),
			).withURLHint(
				makeSearchURL(schemaName, version,
					fmt.Sprintf("CREATE TABLE %s", table.GetName()),
					fmt.Sprintf("ALTER TABLE ONLY %s", table.GetName()),
				),
			)
		}

		if alterStatements, ok := (*column).AlterToTarget(table, expectedColumn); ok {
			return newDriftSummary(
				expectedColumn.GetName(),
				fmt.Sprintf("Unexpected properties of column %s.%q", table.GetName(), expectedColumn.GetName()),
				"alter the column",
			).withStatements(
				alterStatements...,
			)
		}

		return newDriftSummary(
			fmt.Sprintf("%q.%q", table.GetName(), expectedColumn.GetName()),
			fmt.Sprintf("Unexpected properties of column %q.%q", table.GetName(), expectedColumn.GetName()),
			"redefine the column",
		).withDiff(
			expectedColumn,
			*column,
		).withURLHint(
			makeSearchURL(schemaName, version,
				fmt.Sprintf("CREATE TABLE %s", table.GetName()),
				fmt.Sprintf("ALTER TABLE ONLY %s", table.GetName()),
			),
		)
	}
}

func compareColumnsAdditionalCallbackFor(table schemas.TableDescription) func(_ []schemas.ColumnDescription) []Summary {
	return func(additional []schemas.ColumnDescription) []Summary {
		summaries := []Summary{}
		for _, column := range additional {
			summaries = append(summaries, newDriftSummary(
				fmt.Sprintf("%q.%q", table.GetName(), column.GetName()),
				fmt.Sprintf("Unexpected column %q.%q", table.GetName(), column.GetName()),
				"drop the column",
			).withStatements(
				column.DropStatement(table),
			))
		}

		return summaries
	}
}
