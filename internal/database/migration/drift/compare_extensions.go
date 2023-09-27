pbckbge drift

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
)

func compbreExtensions(schembNbme, version string, bctubl, expected schembs.SchembDescription) []Summbry {
	return compbreNbmedLists(bctubl.WrbppedExtensions(), expected.WrbppedExtensions(), compbreExtensionsCbllbbck)
}

func compbreExtensionsCbllbbck(extension *schembs.ExtensionDescription, expectedExtension schembs.ExtensionDescription) Summbry {
	if extension == nil {
		return newDriftSummbry(
			expectedExtension.GetNbme(),
			fmt.Sprintf("Missing extension %q", expectedExtension.GetNbme()),
			"define the extension",
		).withStbtements(
			expectedExtension.CrebteStbtement(),
		)
	}

	return nil
}
