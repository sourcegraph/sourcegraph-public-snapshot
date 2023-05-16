package drift

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func compareSequences(schemaName, version string, actual, expected schemas.SchemaDescription) []Summary {
	return compareNamedLists(
		actual.Sequences,
		expected.Sequences,
		compareSequencesCallbackFor(schemaName, version),
		noopAdditionalCallback[schemas.SequenceDescription],
	)
}

func compareSequencesCallbackFor(schemaName, version string) func(_ *schemas.SequenceDescription, _ schemas.SequenceDescription) Summary {
	return func(sequence *schemas.SequenceDescription, expectedSequence schemas.SequenceDescription) Summary {
		definitionStmt := makeSearchURL(schemaName, version,
			fmt.Sprintf("CREATE SEQUENCE %s", expectedSequence.Name),
			fmt.Sprintf("nextval('%s'::regclass);", expectedSequence.Name),
		)

		if sequence == nil {
			return newDriftSummary(
				expectedSequence.Name,
				fmt.Sprintf("Missing sequence %q", expectedSequence.Name),
				"define the sequence",
			).withURLHint(definitionStmt)
		}

		return newDriftSummary(
			expectedSequence.Name,
			fmt.Sprintf("Unexpected properties of sequence %q", expectedSequence.Name),
			"redefine the sequence",
		).withDiff(expectedSequence, *sequence).withURLHint(definitionStmt)
	}
}
