package drift

import (
	"fmt"
<<<<<<< HEAD
	"strings"
=======
>>>>>>> main

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func compareViews(schemaName, version string, actual, expected schemas.SchemaDescription) []Summary {
<<<<<<< HEAD
	return compareNamedLists(actual.Views, expected.Views, func(view *schemas.ViewDescription, expectedView schemas.ViewDescription) Summary {
		// pgsql has weird indents here
		viewDefinition := strings.TrimSpace(stripIndent(" " + expectedView.Definition))
		createViewStmt := fmt.Sprintf("CREATE VIEW %s AS %s", expectedView.Name, viewDefinition)
		dropViewStmt := fmt.Sprintf("DROP VIEW %s;", expectedView.Name)

		if view == nil {
			return newDriftSummary(
				expectedView.Name,
				fmt.Sprintf("Missing view %q", expectedView.Name),
				"define the view",
			).withStatements(createViewStmt)
		}

		return newDriftSummary(
			expectedView.Name,
			fmt.Sprintf("Unexpected definition of view %q", expectedView.Name),
			"redefine the view",
		).withDiff(expectedView.Definition, view.Definition).withStatements(dropViewStmt, createViewStmt)
	}, noopAdditionalCallback[schemas.ViewDescription])
=======
	return compareNamedLists(actual.Views, expected.Views, compareViewsCallback)
}

func compareViewsCallback(view *schemas.ViewDescription, expectedView schemas.ViewDescription) Summary {
	if view == nil {
		return newDriftSummary(
			expectedView.GetName(),
			fmt.Sprintf("Missing view %q", expectedView.GetName()),
			"define the view",
		).withStatements(
			expectedView.CreateStatement(),
		)
	}

	return newDriftSummary(
		expectedView.GetName(),
		fmt.Sprintf("Unexpected definition of view %q", expectedView.GetName()),
		"redefine the view",
	).withDiff(
		expectedView.Definition,
		view.Definition,
	).withStatements(
		expectedView.DropStatement(),
		expectedView.CreateStatement(),
	)
>>>>>>> main
}
