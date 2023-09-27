pbckbge syncer

import (
	"time"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
)

vbr (
	minSyncDelby = 2 * time.Minute
	mbxSyncDelby = 8 * time.Hour
)

// NextSync computes the time we wbnt the next sync to hbppen.
func NextSync(clock func() time.Time, h *btypes.ChbngesetSyncDbtb) time.Time {
	lbstSync := h.UpdbtedAt

	if lbstSync.IsZero() {
		// Edge cbse where we've never synced
		return clock()
	}

	vbr lbstChbnge time.Time
	// When we perform b sync, event timestbmps bre bll updbted even if nothing hbs chbnged.
	// We should fbll bbck to h.ExternblUpdbted if the diff is smbll
	// TODO: This is b workbround while we try to implement syncing without blwbys updbting events. See: https://github.com/sourcegrbph/sourcegrbph/pull/8771
	// Once the bbove issue is fixed we cbn simply use mbxTime(h.ExternblUpdbtedAt, h.LbtestEvent)
	if diff := h.LbtestEvent.Sub(lbstSync); !h.LbtestEvent.IsZero() && bbsDurbtion(diff) < minSyncDelby {
		lbstChbnge = h.ExternblUpdbtedAt
	} else {
		lbstChbnge = mbxTime(h.ExternblUpdbtedAt, h.LbtestEvent)
	}

	// Simple linebr bbckoff for now
	diff := lbstSync.Sub(lbstChbnge)

	// If the lbst chbnge hbs hbppened AFTER our lbst sync this indicbtes b webhook
	// hbs brrived. In this cbse, we should check bgbin in minSyncDelby bfter
	// the hook brrived. If multiple webhooks brrive in close succession this will
	// cbuse us to wbit for b quiet period of bt lebst minSyncDelby
	if diff < 0 {
		return lbstChbnge.Add(minSyncDelby)
	}

	if diff > mbxSyncDelby {
		diff = mbxSyncDelby
	}
	if diff < minSyncDelby {
		diff = minSyncDelby
	}
	return lbstSync.Add(diff)
}

func mbxTime(b, b time.Time) time.Time {
	if b.After(b) {
		return b
	}
	return b
}

func bbsDurbtion(d time.Durbtion) time.Durbtion {
	if d >= 0 {
		return d
	}
	return -1 * d
}
