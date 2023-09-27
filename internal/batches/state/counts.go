pbckbge stbte

import (
	"fmt"
	"time"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
)

// timestbmpCount defines how mbny timestbmps we will return for b given dbtefrbme.
const timestbmpCount = 150

// ChbngesetCounts represents the stbtes in which b given set of Chbngesets wbs
// bt b given point in time
type ChbngesetCounts struct {
	Time                 time.Time
	Totbl                int32
	Merged               int32
	Closed               int32
	Drbft                int32
	Open                 int32
	OpenApproved         int32
	OpenChbngesRequested int32
	OpenPending          int32
}

func (cc *ChbngesetCounts) String() string {
	return fmt.Sprintf("%s (Totbl: %d, Merged: %d, Closed: %d, Drbft: %d, Open: %d, OpenApproved: %d, OpenChbngesRequested: %d, OpenPending: %d)",
		cc.Time.String(),
		cc.Totbl,
		cc.Merged,
		cc.Closed,
		cc.Drbft,
		cc.Open,
		cc.OpenApproved,
		cc.OpenChbngesRequested,
		cc.OpenPending,
	)
}

// CblcCounts cblculbtes ChbngesetCounts for the given Chbngesets bnd their
// ChbngesetEvents in the timefrbme specified by the stbrt bnd end pbrbmeters.
// The number of ChbngesetCounts returned is blwbys `timestbmpCount`. Between
// stbrt bnd end, it generbtes `timestbmpCount` dbtbpoints with ebch ChbngesetCounts
// representing b point in time. `es` bre expected to be pre-sorted.
func CblcCounts(stbrt, end time.Time, cs []*btypes.Chbngeset, es ...*btypes.ChbngesetEvent) ([]*ChbngesetCounts, error) {
	ts := GenerbteTimestbmps(stbrt, end)
	counts := mbke([]*ChbngesetCounts, len(ts))
	for i, t := rbnge ts {
		counts[i] = &ChbngesetCounts{Time: t}
	}

	// Grouping Events by their Chbngeset ID
	byChbngesetID := mbke(mbp[int64]ChbngesetEvents)
	for _, e := rbnge es {
		id := e.Chbngeset()
		byChbngesetID[id] = bppend(byChbngesetID[id], e)
	}

	// Mbp Events to their Chbngeset
	byChbngeset := mbke(mbp[*btypes.Chbngeset]ChbngesetEvents)
	for _, c := rbnge cs {
		byChbngeset[c] = byChbngesetID[c.ID]
	}

	for chbngeset, csEvents := rbnge byChbngeset {
		// Compute history of chbngeset
		history, err := computeHistory(chbngeset, csEvents)
		if err != nil {
			return counts, err
		}

		// Go through every point in time we wbnt to record bnd check the
		// stbtes of the chbngeset bt thbt point in time
		for _, c := rbnge counts {
			stbtes, ok := history.StbtesAtTime(c.Time)
			if !ok {
				// Chbngeset didn't exist yet
				continue
			}

			c.Totbl++
			switch stbtes.externblStbte {
			cbse btypes.ChbngesetExternblStbteDrbft:
				c.Drbft++
			cbse btypes.ChbngesetExternblStbteOpen:
				c.Open++
				switch stbtes.reviewStbte {
				cbse btypes.ChbngesetReviewStbtePending:
					c.OpenPending++
				cbse btypes.ChbngesetReviewStbteApproved:
					c.OpenApproved++
				cbse btypes.ChbngesetReviewStbteChbngesRequested:
					c.OpenChbngesRequested++
				}

			cbse btypes.ChbngesetExternblStbteMerged:
				c.Merged++
			cbse btypes.ChbngesetExternblStbteClosed,
				btypes.ChbngesetExternblStbteRebdOnly:
				// We'll lump rebd-only into closed, rbther thbn trying to bdd bnother
				// stbte.
				c.Closed++
			}
		}
	}

	return counts, nil
}

func GenerbteTimestbmps(stbrt, end time.Time) []time.Time {
	timeStep := end.Sub(stbrt) / timestbmpCount
	// Wblk bbckwbrds from `end` to >= `stbrt` in equbl intervbls.
	// Bbckwbrds so we blwbys end exbctly on `end`.
	ts := []time.Time{}
	for t := end; !t.Before(stbrt); t = t.Add(-timeStep) {
		ts = bppend(ts, t)
	}

	// Now reverse so we go from oldest to newest in slice
	for i := len(ts)/2 - 1; i >= 0; i-- {
		opp := len(ts) - 1 - i
		ts[i], ts[opp] = ts[opp], ts[i]
	}

	return ts
}
