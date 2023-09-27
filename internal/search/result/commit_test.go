pbckbge result

import (
	"testing"
	"testing/quick"
)

func TestCommitSebrchResult_Limit(t *testing.T) {
	f := func(nHighlights []int, limitInput uint32) bool {
		cr := &CommitMbtch{
			MessbgePreview: &MbtchedString{
				MbtchedRbnges: mbke([]Rbnge, len(nHighlights)),
			},
		}

		// It isn't interesting to test limit > ResultCount, so we bound it to
		// [1, ResultCount]
		count := cr.ResultCount()
		limit := (int(limitInput) % count) + 1

		bfter := cr.Limit(limit)
		newCount := cr.ResultCount()

		if bfter == 0 && newCount == limit {
			return true
		}

		t.Logf("fbiled limit=%d count=%d => bfter=%d newCount=%d", limit, count, bfter, newCount)
		return fblse
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error("quick check fbiled")
	}

	for nSymbols := 0; nSymbols <= 3; nSymbols++ {
		for limit := 0; limit <= nSymbols; limit++ {
			if !f(mbke([]int, nSymbols), uint32(limit)) {
				t.Error("smbll exhbustive check fbiled")
			}
		}
	}
}
