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

func compareIndexesCallbackFor(expectedTable schemas.TableDescription) func(_ *schemas.IndexDescription, _ schemas.IndexDescription) Summary {
	return func(index *schemas.IndexDescription, expectedIndex schemas.IndexDescription) Summary {
		var createIndexStmt string
		switch expectedIndex.ConstraintType {
		case "u":
			fallthrough
		case "p":
			createIndexStmt = fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s %s;", expectedTable.Name, expectedIndex.Name, expectedIndex.ConstraintDefinition)
		default:
			createIndexStmt = fmt.Sprintf("%s;", expectedIndex.IndexDefinition)
		}
		dropIndexStmt := fmt.Sprintf("DROP INDEX %s;", expectedIndex.Name)

		if index == nil {
			return newDriftSummary(
				fmt.Sprintf("%q.%q", expectedTable.Name, expectedIndex.Name),
				fmt.Sprintf("Missing index %q.%q", expectedTable.Name, expectedIndex.Name),
				"define the index",
			).withStatements(createIndexStmt)
		}

		return newDriftSummary(
			fmt.Sprintf("%q.%q", expectedTable.Name, expectedIndex.Name),
			fmt.Sprintf("Unexpected properties of index %q.%q", expectedTable.Name, expectedIndex.Name),
			"redefine the index",
		).withDiff(expectedIndex, *index).withStatements(dropIndexStmt, createIndexStmt)
	}
}

func compareIndexesCallbackAdditionalFor(expectedTable schemas.TableDescription) func(_ []schemas.IndexDescription) []Summary {
	return func(additional []schemas.IndexDescription) []Summary {
		summaries := []Summary{}
		for _, index := range additional {
			dropIndexStmt := fmt.Sprintf("DROP INDEX %s;", index.Name)

			summary := newDriftSummary(
				fmt.Sprintf("%q.%q", expectedTable.Name, index.Name),
				fmt.Sprintf("Unexpected index %q.%q", expectedTable.Name, index.Name),
				"drop the index",
			).withStatements(dropIndexStmt)
			summaries = append(summaries, summary)
		}

		return summaries
	}
}
