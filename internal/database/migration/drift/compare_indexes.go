package drift

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func compareIndexes(actualTable, expectedTable schemas.TableDescription) []Summary {
	return compareNamedListsStrict(
		actualTable.Indexes,
		expectedTable.Indexes,
		compareIndexesCallbackFor(expectedTable),
		compareIndexesCallbackAdditionalFor(expectedTable),
	)
}

func compareIndexesCallbackFor(table schemas.TableDescription) func(_ *schemas.IndexDescription, _ schemas.IndexDescription) Summary {
	return func(index *schemas.IndexDescription, expectedIndex schemas.IndexDescription) Summary {
		if index == nil {
			return newDriftSummary(
				fmt.Sprintf("%q.%q", table.GetName(), expectedIndex.GetName()),
				fmt.Sprintf("Missing index %q.%q", table.GetName(), expectedIndex.GetName()),
				"define the index",
			).withStatements(
				expectedIndex.CreateStatement(table),
			)
		}

		return newDriftSummary(
			fmt.Sprintf("%q.%q", table.GetName(), expectedIndex.GetName()),
			fmt.Sprintf("Unexpected properties of index %q.%q", table.GetName(), expectedIndex.GetName()),
			"redefine the index",
		).withDiff(
			expectedIndex,
			*index,
		).withStatements(
			expectedIndex.DropStatement(table),
			expectedIndex.CreateStatement(table),
		)
	}
}

func compareIndexesCallbackAdditionalFor(table schemas.TableDescription) func(_ []schemas.IndexDescription) []Summary {
	return func(additional []schemas.IndexDescription) []Summary {
		summaries := []Summary{}
		for _, index := range additional {
			summaries = append(summaries, newDriftSummary(
				fmt.Sprintf("%q.%q", table.GetName(), index.GetName()),
				fmt.Sprintf("Unexpected index %q.%q", table.GetName(), index.GetName()),
				"drop the index",
			).withStatements(
				index.DropStatement(table),
			))
		}

		return summaries
	}
}
