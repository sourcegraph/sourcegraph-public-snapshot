package drift

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func compareConstraints(actualTable, expectedTable schemas.TableDescription) []Summary {
	return compareNamedLists(actualTable.Constraints, expectedTable.Constraints, func(constraint *schemas.ConstraintDescription, expectedConstraint schemas.ConstraintDescription) Summary {
		createConstraintStmt := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s %s;", expectedTable.Name, expectedConstraint.Name, expectedConstraint.ConstraintDefinition)
		dropConstraintStmt := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s;", expectedTable.Name, expectedConstraint.Name)

		if constraint == nil {
			return newDriftSummary(
				fmt.Sprintf("%q.%q", expectedTable.Name, expectedConstraint.Name),
				fmt.Sprintf("Missing constraint %q.%q", expectedTable.Name, expectedConstraint.Name),
				"define the constraint",
			).withStatements(createConstraintStmt)
		}

		return newDriftSummary(
			fmt.Sprintf("%q.%q", expectedTable.Name, expectedConstraint.Name),
			fmt.Sprintf("Unexpected properties of constraint %q.%q", expectedTable.Name, expectedConstraint.Name),
			"redefine the constraint",
		).withDiff(expectedConstraint, *constraint).withStatements(dropConstraintStmt, createConstraintStmt)
	}, func(additional []schemas.ConstraintDescription) []Summary {
		summaries := []Summary{}
		for _, constraint := range additional {
			alterTableStmt := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s;", expectedTable.Name, constraint.Name)

			summary := newDriftSummary(
				fmt.Sprintf("%q.%q", expectedTable.Name, constraint.Name),
				fmt.Sprintf("Unexpected constraint %q.%q", expectedTable.Name, constraint.Name),
				"drop the constraint",
			).withStatements(alterTableStmt)
			summaries = append(summaries, summary)
		}

		return summaries
	})
}
