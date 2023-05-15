package drift

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func compareTables(schemaName, version string, actual, expected schemas.SchemaDescription) []Summary {
	return compareNamedListsMulti(actual.Tables, expected.Tables, func(table *schemas.TableDescription, expectedTable schemas.TableDescription) []Summary {
		if table == nil {
			url := makeSearchURL(schemaName, version,
				fmt.Sprintf("CREATE TABLE %s", expectedTable.Name),
				fmt.Sprintf("ALTER TABLE ONLY %s", expectedTable.Name),
				fmt.Sprintf("CREATE .*(INDEX|TRIGGER).* ON %s", expectedTable.Name),
			)

			return singleton(newDriftSummary(
				expectedTable.Name,
				fmt.Sprintf("Missing table %q", expectedTable.Name),
				"define the table",
			).withURLHint(url))
		}

		summaries := []Summary(nil)
		summaries = append(summaries, compareColumns(schemaName, version, *table, expectedTable)...)
		summaries = append(summaries, compareConstraints(*table, expectedTable)...)
		summaries = append(summaries, compareIndexes(*table, expectedTable)...)
		summaries = append(summaries, compareTriggers(*table, expectedTable)...)
		return summaries
	}, noopAdditionalCallback[schemas.TableDescription])
}
