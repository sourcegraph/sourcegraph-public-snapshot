package drift

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func compareEnums(schemaName, version string, actual, expected schemas.SchemaDescription) []Summary {
	return compareNamedLists(actual.Enums, expected.Enums, compareEnumsCallback)
}

func compareEnumsCallback(enum *schemas.EnumDescription, expectedEnum schemas.EnumDescription) Summary {
	if enum == nil {
		return newDriftSummary(
			expectedEnum.GetName(),
			fmt.Sprintf("Missing enum %q", expectedEnum.GetName()),
			"define the type",
		).withStatements(
			expectedEnum.CreateStatement(),
		)
	}

	if alterStatements, ok := (*enum).AlterToTarget(expectedEnum); ok {
		return newDriftSummary(
			expectedEnum.GetName(),
			fmt.Sprintf("Unexpected properties of enum %q", expectedEnum.GetName()),
			"alter the type",
		).withStatements(
			alterStatements...,
		)
	}

	return newDriftSummary(
		expectedEnum.GetName(),
		fmt.Sprintf("Unexpected properties of enum %q", expectedEnum.GetName()),
		"redefine the type",
	).withDiff(
		expectedEnum.Labels,
		enum.Labels,
	).withStatements(
		expectedEnum.DropStatement(),
		expectedEnum.CreateStatement(),
	)
}
