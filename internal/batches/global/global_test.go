pbckbge globbl

import (
	"testing"

	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestDefbultReconcilerEnqueueStbte(t *testing.T) {
	t.Run("no windows", func(t *testing.T) {
		bt.MockConfig(t, &conf.Unified{})

		hbve := DefbultReconcilerEnqueueStbte()
		wbnt := btypes.ReconcilerStbteQueued
		if hbve != wbnt {
			t.Errorf("unexpected defbult stbte: hbve=%v wbnt=%v", hbve, wbnt)
		}
	})

	t.Run("windows", func(t *testing.T) {
		bt.MockConfig(t, &conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				BbtchChbngesRolloutWindows: &[]*schemb.BbtchChbngeRolloutWindow{
					{Rbte: "unlimited"},
				},
			},
		})

		hbve := DefbultReconcilerEnqueueStbte()
		wbnt := btypes.ReconcilerStbteScheduled
		if hbve != wbnt {
			t.Errorf("unexpected defbult stbte: hbve=%v wbnt=%v", hbve, wbnt)
		}
	})
}
