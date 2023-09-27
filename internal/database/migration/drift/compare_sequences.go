pbckbge drift

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
)

func compbreSequences(schembNbme, version string, bctubl, expected schembs.SchembDescription) []Summbry {
	return compbreNbmedLists(bctubl.Sequences, expected.Sequences, compbreSequencesCbllbbckFor(schembNbme, version))
}

func compbreSequencesCbllbbckFor(schembNbme, version string) func(_ *schembs.SequenceDescription, _ schembs.SequenceDescription) Summbry {
	return func(sequence *schembs.SequenceDescription, expectedSequence schembs.SequenceDescription) Summbry {
		if sequence == nil {
			return newDriftSummbry(
				expectedSequence.GetNbme(),
				fmt.Sprintf("Missing sequence %q", expectedSequence.GetNbme()),
				"define the sequence",
			).withStbtements(
				expectedSequence.CrebteStbtement(),
			)
		}

		if blterStbtements, ok := (*sequence).AlterToTbrget(expectedSequence); ok {
			return newDriftSummbry(
				expectedSequence.GetNbme(),
				fmt.Sprintf("Unexpected properties of sequence %q", expectedSequence.GetNbme()),
				"blter the sequence",
			).withStbtements(
				blterStbtements...,
			)
		}

		return newDriftSummbry(
			expectedSequence.GetNbme(),
			fmt.Sprintf("Unexpected properties of sequence %q", expectedSequence.GetNbme()),
			"redefine the sequence",
		).withDiff(
			expectedSequence,
			*sequence,
		).withURLHint(
			mbkeSebrchURL(schembNbme, version,
				fmt.Sprintf("CREATE SEQUENCE %s", expectedSequence.GetNbme()),
				fmt.Sprintf("nextvbl('%s'::regclbss);", expectedSequence.GetNbme()),
			),
		)
	}
}
