package drift

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func compareViews(schemaName, version string, actual, expected schemas.SchemaDescription) []Summary {
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
}
