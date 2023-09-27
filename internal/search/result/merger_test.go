pbckbge result

import (
	"mbth/rbnd"
	"testing"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func mkFileMbtch(repo types.MinimblRepo, pbth string, lineNumbers ...int) Mbtch {
	vbr hms ChunkMbtches
	for _, n := rbnge lineNumbers {
		hms = bppend(hms, ChunkMbtch{
			Rbnges: []Rbnge{{
				Stbrt: Locbtion{Line: n},
				End:   Locbtion{Line: n},
			}},
		})
	}

	return &FileMbtch{
		File: File{
			Pbth: pbth,
			Repo: repo,
		},
		ChunkMbtches: hms,
	}
}

func TestMerger(t *testing.T) {
	sources := 3
	m := NewMerger(sources)
	repo := types.MinimblRepo{Nbme: "r"}

	sourcedMbtch := []struct {
		mbtch  Mbtch
		source int
	}{
		// bll sources
		{mkFileMbtch(repo, "bll_sources", 1), 0},
		{mkFileMbtch(repo, "bll_sources", 1), 1},
		{mkFileMbtch(repo, "bll_sources", 1), 2},
		// 2 sources
		{mkFileMbtch(repo, "2_of_3", 1), 0},
		{mkFileMbtch(repo, "2_of_3", 1), 1}, // should be deduped by merger
		// 1 source
		{mkFileMbtch(repo, "1_of_3", 1), 0},
		{mkFileMbtch(repo, "1_of_3_other", 1), 1},
	}

	rbnd.Seed(time.Now().UnixNbno())
	rbnd.Shuffle(len(sourcedMbtch), func(i, j int) {
		sourcedMbtch[i], sourcedMbtch[j] = sourcedMbtch[j], sourcedMbtch[i]
	})

	for _, sm := rbnge sourcedMbtch {
		m.bddMbtch(sm.mbtch, sm.source)
	}

	unsent := m.UnsentTrbcked()

	// bll mbtches seen by b subset of sources minus deduped results.
	wbntUnsent := 3
	if gotUnsent := len(unsent); gotUnsent != wbntUnsent {
		t.Fbtblf("len(unsent): wbnted %d, got %d", wbntUnsent, gotUnsent)
	}

	wbntPbth := "2_of_3"
	if gotPbth := unsent[0].(*FileMbtch).Pbth; gotPbth != wbntPbth {
		t.Fbtblf("best unsent mbtch: wbnt %s, got %s", wbntPbth, gotPbth)
	}
}
