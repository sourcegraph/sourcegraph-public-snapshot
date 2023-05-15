package drift

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func compareSequences(schemaName, version string, actual, expected schemas.SchemaDescription) []Summary {
	return compareNamedLists(actual.Sequences, expected.Sequences, func(sequence *schemas.SequenceDescription, expectedSequence schemas.SequenceDescription) Summary {
		definitionStmt := makeSearchURL(schemaName, version,
			fmt.Sprintf("CREATE SEQUENCE %s", expectedSequence.Name),
			fmt.Sprintf("nextval('%s'::regclass);", expectedSequence.Name),
		)

		if sequence == nil {
			return newDriftSummary(
				expectedSequence.Name,
				fmt.Sprintf("Missing sequence %q", expectedSequence.Name),
				"define the sequence",
			).withStatements(definitionStmt)
		}

		return newDriftSummary(
			expectedSequence.Name,
			fmt.Sprintf("Unexpected properties of sequence %q", expectedSequence.Name),
			"redefine the sequence",
		).withDiff(expectedSequence, *sequence).withStatements(definitionStmt)
	}, noopAdditionalCallback[schemas.SequenceDescription])
}
