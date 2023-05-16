package drift

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func compareFunctions(schemaName, version string, actual, expected schemas.SchemaDescription) []Summary {
	return compareNamedLists(actual.Functions, expected.Functions, compareFunctionsCallback)
}

func compareFunctionsCallback(function *schemas.FunctionDescription, expectedFunction schemas.FunctionDescription) Summary {
	if function == nil {
		return newDriftSummary(
			expectedFunction.GetName(),
			fmt.Sprintf("Missing function %q", expectedFunction.GetName()),
			"define the function",
		).withStatements(
			expectedFunction.CreateOrReplaceStatement(),
		)
	}

	return newDriftSummary(
		expectedFunction.GetName(),
		fmt.Sprintf("Unexpected definition of function %q", expectedFunction.GetName()),
		"redefine the function",
	).withDiff(
		expectedFunction.Definition,
		function.Definition,
	).withStatements(
		expectedFunction.CreateOrReplaceStatement(),
	)
}
