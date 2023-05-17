package drift

import (
	"fmt"
<<<<<<< HEAD
	"strings"
=======
>>>>>>> main

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func compareFunctions(schemaName, version string, actual, expected schemas.SchemaDescription) []Summary {
<<<<<<< HEAD
	return compareNamedLists(actual.Functions, expected.Functions, func(function *schemas.FunctionDescription, expectedFunction schemas.FunctionDescription) Summary {
		definitionStmt := fmt.Sprintf("%s;", strings.TrimSpace(expectedFunction.Definition))

		if function == nil {
			return newDriftSummary(
				expectedFunction.Name,
				fmt.Sprintf("Missing function %q", expectedFunction.Name),
				"define the function",
			).withStatements(definitionStmt)
		}

		return newDriftSummary(
			expectedFunction.Name,
			fmt.Sprintf("Unexpected definition of function %q", expectedFunction.Name),
			"replace the function definition",
		).withDiff(expectedFunction.Definition, function.Definition).withStatements(definitionStmt)
	}, noopAdditionalCallback[schemas.FunctionDescription])
=======
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
>>>>>>> main
}
