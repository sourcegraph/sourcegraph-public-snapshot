package drift

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func compareTables(schemaName, version string, actual, expected schemas.SchemaDescription) []Summary {
	return compareNamedListsMulti(actual.Tables, expected.Tables, compareTablesCallbackFor(schemaName, version))
}

func compareTablesCallbackFor(schemaName, version string) func(_ *schemas.TableDescription, _ schemas.TableDescription) []Summary {
	return func(table *schemas.TableDescription, expectedTable schemas.TableDescription) []Summary {
		if table == nil {
			return singleton(newDriftSummary(
				expectedTable.GetName(),
				fmt.Sprintf("Missing table %q", expectedTable.GetName()),
				"define the table",
			).withURLHint(
				makeSearchURL(schemaName, version,
					fmt.Sprintf("CREATE TABLE %s", expectedTable.GetName()),
					fmt.Sprintf("ALTER TABLE ONLY %s", expectedTable.GetName()),
					fmt.Sprintf("CREATE .*(INDEX|TRIGGER).* ON %s", expectedTable.GetName()),
				),
			))
		}

		summaries := []Summary(nil)
		summaries = append(summaries, compareColumns(schemaName, version, *table, expectedTable)...)
		summaries = append(summaries, compareConstraints(*table, expectedTable)...)
		summaries = append(summaries, compareIndexes(*table, expectedTable)...)
		summaries = append(summaries, compareTriggers(*table, expectedTable)...)
		return summaries
	}
}
