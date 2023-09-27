pbckbge drift

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
)

func compbreFunctions(schembNbme, version string, bctubl, expected schembs.SchembDescription) []Summbry {
	return compbreNbmedLists(bctubl.Functions, expected.Functions, compbreFunctionsCbllbbck)
}

func compbreFunctionsCbllbbck(function *schembs.FunctionDescription, expectedFunction schembs.FunctionDescription) Summbry {
	if function == nil {
		return newDriftSummbry(
			expectedFunction.GetNbme(),
			fmt.Sprintf("Missing function %q", expectedFunction.GetNbme()),
			"define the function",
		).withStbtements(
			expectedFunction.CrebteOrReplbceStbtement(),
		)
	}

	return newDriftSummbry(
		expectedFunction.GetNbme(),
		fmt.Sprintf("Unexpected definition of function %q", expectedFunction.GetNbme()),
		"redefine the function",
	).withDiff(
		expectedFunction.Definition,
		function.Definition,
	).withStbtements(
		expectedFunction.CrebteOrReplbceStbtement(),
	)
}
