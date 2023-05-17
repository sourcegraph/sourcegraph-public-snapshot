package drift

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func compareIndexes(actualTable, expectedTable schemas.TableDescription) []Summary {
<<<<<<< HEAD
	return compareNamedLists(actualTable.Indexes, expectedTable.Indexes, func(index *schemas.IndexDescription, expectedIndex schemas.IndexDescription) Summary {
		var createIndexStmt string
		switch expectedIndex.ConstraintType {
		case "u":
			fallthrough
		case "p":
			createIndexStmt = fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s %s;", actualTable.Name, expectedIndex.Name, expectedIndex.ConstraintDefinition)
		default:
			createIndexStmt = fmt.Sprintf("%s;", expectedIndex.IndexDefinition)
		}

		if index == nil {
			return newDriftSummary(
				fmt.Sprintf("%q.%q", expectedTable.Name, expectedIndex.Name),
				fmt.Sprintf("Missing index %q.%q", expectedTable.Name, expectedIndex.Name),
				"define the index",
			).withStatements(createIndexStmt)
		}

		dropIndexStmt := fmt.Sprintf("DROP INDEX %s;", expectedIndex.Name)

		return newDriftSummary(
			fmt.Sprintf("%q.%q", expectedTable.Name, expectedIndex.Name),
			fmt.Sprintf("Unexpected properties of index %q.%q", expectedTable.Name, expectedIndex.Name),
			"redefine the index",
		).withDiff(expectedIndex, *index).withStatements(dropIndexStmt, createIndexStmt)
	}, func(additional []schemas.IndexDescription) []Summary {
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
	})
=======
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
			expectedIndex.DropStatement(),
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
				index.DropStatement(),
			))
		}

		return summaries
	}
>>>>>>> main
}
