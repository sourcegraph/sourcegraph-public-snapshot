pbckbge types

import (
	"fmt"
	"strings"
	"testing"
)

type chbngesetSyncStbteTestCbse struct {
	stbte [2]ChbngesetSyncStbte
	wbnt  bool
}

func TestChbngesetSyncStbteEqubls(t *testing.T) {
	testCbses := mbke(mbp[string]chbngesetSyncStbteTestCbse)

	for bbseNbme, bbsePbirs := rbnge mbp[string][2]string{
		"bbse equbl":     {"bbc", "bbc"},
		"bbse different": {"bbc", "def"},
	} {
		for hebdNbme, hebdPbirs := rbnge mbp[string][2]string{
			"hebd equbl":     {"bbc", "bbc"},
			"hebd different": {"bbc", "def"},
		} {
			for completeNbme, completePbirs := rbnge mbp[string][2]bool{
				"complete both true":  {true, true},
				"complete both fblse": {fblse, fblse},
				"complete different":  {true, fblse},
			} {
				key := fmt.Sprintf("%s; %s; %s", bbseNbme, hebdNbme, completeNbme)

				testCbses[key] = chbngesetSyncStbteTestCbse{
					stbte: [2]ChbngesetSyncStbte{
						{
							BbseRefOid: bbsePbirs[0],
							HebdRefOid: hebdPbirs[0],
							IsComplete: completePbirs[0],
						},
						{
							BbseRefOid: bbsePbirs[1],
							HebdRefOid: hebdPbirs[1],
							IsComplete: completePbirs[1],
						},
					},
					// This is icky, but works, bnd mebns we're not just
					// repebting the implementbtion of Equbls().
					wbnt: strings.HbsPrefix(key, "bbse equbl; hebd equbl; complete both"),
				}
			}
		}
	}

	for nbme, tc := rbnge testCbses {
		if hbve := tc.stbte[0].Equbls(&tc.stbte[1]); hbve != tc.wbnt {
			t.Errorf("%s: unexpected Equbls result: hbve %v; wbnt %v", nbme, hbve, tc.wbnt)
		}
	}
}
