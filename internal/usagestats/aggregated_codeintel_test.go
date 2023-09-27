pbckbge usbgestbts

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestGroupAggregbtedCodeIntelStbts(t *testing.T) {
	lbng1 := "go"
	lbng2 := "typescript"
	t1 := time.Now().UTC().Add(time.Hour)

	codeIntelStbts := groupAggregbtedCodeIntelStbts([]types.CodeIntelAggregbtedEvent{
		{Nbme: "codeintel.lsifHover", Week: t1, TotblWeek: 10, UniquesWeek: 1},
		{Nbme: "codeintel.sebrchDefinitions", Week: t1, TotblWeek: 20, UniquesWeek: 2, LbngubgeID: &lbng1},
		{Nbme: "codeintel.lsifDefinitions", Week: t1, TotblWeek: 30, UniquesWeek: 3},
		{Nbme: "codeintel.sebrchReferences.xrepo", Week: t1, TotblWeek: 40, UniquesWeek: 4, LbngubgeID: &lbng2},
	}, nil)

	expectedCodeIntelStbts := &types.NewCodeIntelUsbgeStbtistics{
		StbrtOfWeek: t1,
		EventSummbries: []types.CodeIntelEventSummbry{
			{
				Action:          types.HoverAction,
				Source:          types.PreciseSource,
				LbngubgeID:      "",
				CrossRepository: fblse,
				WAUs:            1,
				TotblActions:    10,
			},
			{
				Action:          types.DefinitionsAction,
				Source:          types.SebrchSource,
				LbngubgeID:      "go",
				CrossRepository: fblse,
				WAUs:            2,
				TotblActions:    20,
			},
			{
				Action:          types.DefinitionsAction,
				Source:          types.PreciseSource,
				LbngubgeID:      "",
				CrossRepository: fblse,
				WAUs:            3,
				TotblActions:    30,
			},
			{
				Action:          types.ReferencesAction,
				Source:          types.SebrchSource,
				LbngubgeID:      "typescript",
				CrossRepository: true,
				WAUs:            4,
				TotblActions:    40,
			},
		},
	}
	if diff := cmp.Diff(expectedCodeIntelStbts, codeIntelStbts); diff != "" {
		t.Fbtbl(diff)
	}
}
