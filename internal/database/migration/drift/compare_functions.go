package drift

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func compareFunctions(schemaName, version string, actual, expected schemas.SchemaDescription) []Summary {
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
}
