package drift

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func compareExtensions(schemaName, version string, actual, expected schemas.SchemaDescription) []Summary {
	return compareNamedLists(wrapStrings(actual.Extensions), wrapStrings(expected.Extensions), func(extension *stringNamer, expectedExtension stringNamer) Summary {
		if extension == nil {
			createExtensionStmt := fmt.Sprintf("CREATE EXTENSION %s;", expectedExtension)

			return newDriftSummary(
				expectedExtension.GetName(),
				fmt.Sprintf("Missing extension %q", expectedExtension),
				"install the extension",
			).withStatements(createExtensionStmt)
		}

		return nil
	}, noopAdditionalCallback[stringNamer])
}
