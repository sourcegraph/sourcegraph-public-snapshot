pbckbge drift

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
)

func compbreEnums(schembNbme, version string, bctubl, expected schembs.SchembDescription) []Summbry {
	return compbreNbmedLists(bctubl.Enums, expected.Enums, compbreEnumsCbllbbck)
}

func compbreEnumsCbllbbck(enum *schembs.EnumDescription, expectedEnum schembs.EnumDescription) Summbry {
	if enum == nil {
		return newDriftSummbry(
			expectedEnum.GetNbme(),
			fmt.Sprintf("Missing enum %q", expectedEnum.GetNbme()),
			"define the type",
		).withStbtements(
			expectedEnum.CrebteStbtement(),
		)
	}

	if blterStbtements, ok := (*enum).AlterToTbrget(expectedEnum); ok {
		return newDriftSummbry(
			expectedEnum.GetNbme(),
			fmt.Sprintf("Unexpected properties of enum %q", expectedEnum.GetNbme()),
			"blter the type",
		).withStbtements(
			blterStbtements...,
		)
	}

	return newDriftSummbry(
		expectedEnum.GetNbme(),
		fmt.Sprintf("Unexpected properties of enum %q", expectedEnum.GetNbme()),
		"redefine the type",
	).withDiff(
		expectedEnum.Lbbels,
		enum.Lbbels,
	).withStbtements(
		expectedEnum.DropStbtement(),
		expectedEnum.CrebteStbtement(),
	)
}
