package drift

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func compareViews(schemaName, version string, actual, expected schemas.SchemaDescription) []Summary {
	return compareNamedLists(
		actual.Views,
		expected.Views,
		compareViewsCallback,
		noopAdditionalCallback[schemas.ViewDescription],
	)
}

func compareViewsCallback(view *schemas.ViewDescription, expectedView schemas.ViewDescription) Summary {
	viewDefinition := strings.TrimSpace(stripIndent(" " + expectedView.Definition)) // pgsql has weird indents here
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
}
