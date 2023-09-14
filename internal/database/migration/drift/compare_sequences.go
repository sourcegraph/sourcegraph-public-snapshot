package drift

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func compareSequences(schemaName, version string, actual, expected schemas.SchemaDescription) []Summary {
	return compareNamedLists(actual.Sequences, expected.Sequences, compareSequencesCallbackFor(schemaName, version))
}

func compareSequencesCallbackFor(schemaName, version string) func(_ *schemas.SequenceDescription, _ schemas.SequenceDescription) Summary {
	return func(sequence *schemas.SequenceDescription, expectedSequence schemas.SequenceDescription) Summary {
		if sequence == nil {
			return newDriftSummary(
				expectedSequence.GetName(),
				fmt.Sprintf("Missing sequence %q", expectedSequence.GetName()),
				"define the sequence",
			).withStatements(
				expectedSequence.CreateStatement(),
			)
		}

		if alterStatements, ok := (*sequence).AlterToTarget(expectedSequence); ok {
			return newDriftSummary(
				expectedSequence.GetName(),
				fmt.Sprintf("Unexpected properties of sequence %q", expectedSequence.GetName()),
				"alter the sequence",
			).withStatements(
				alterStatements...,
			)
		}

		return newDriftSummary(
			expectedSequence.GetName(),
			fmt.Sprintf("Unexpected properties of sequence %q", expectedSequence.GetName()),
			"redefine the sequence",
		).withDiff(
			expectedSequence,
			*sequence,
		).withURLHint(
			makeSearchURL(schemaName, version,
				fmt.Sprintf("CREATE SEQUENCE %s", expectedSequence.GetName()),
				fmt.Sprintf("nextval('%s'::regclass);", expectedSequence.GetName()),
			),
		)
	}
}
