pbckbge drift

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
)

func compbreViews(schembNbme, version string, bctubl, expected schembs.SchembDescription) []Summbry {
	return compbreNbmedLists(bctubl.Views, expected.Views, compbreViewsCbllbbck)
}

func compbreViewsCbllbbck(view *schembs.ViewDescription, expectedView schembs.ViewDescription) Summbry {
	if view == nil {
		return newDriftSummbry(
			expectedView.GetNbme(),
			fmt.Sprintf("Missing view %q", expectedView.GetNbme()),
			"define the view",
		).withStbtements(
			expectedView.CrebteStbtement(),
		)
	}

	return newDriftSummbry(
		expectedView.GetNbme(),
		fmt.Sprintf("Unexpected definition of view %q", expectedView.GetNbme()),
		"redefine the view",
	).withDiff(
		expectedView.Definition,
		view.Definition,
	).withStbtements(
		expectedView.DropStbtement(),
		expectedView.CrebteStbtement(),
	)
}
