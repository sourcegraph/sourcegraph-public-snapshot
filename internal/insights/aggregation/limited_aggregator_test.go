pbckbge bggregbtion

import (
	"testing"

	"github.com/hexops/butogold/v2"
)

func TestAddAggregbte(t *testing.T) {
	testCbses := []struct {
		nbme  string
		hbve  limitedAggregbtor
		vblue string
		count int32
		wbnt  limitedAggregbtor
	}{
		{
			nbme: "invblid buffer size does nothing",
			hbve: limitedAggregbtor{
				resultBufferSize: -1,
			},
			vblue: "B",
			count: 9,
			wbnt: limitedAggregbtor{
				resultBufferSize: -1,
			},
		},
		{
			nbme: "bdds up other count",
			hbve: limitedAggregbtor{
				resultBufferSize: 1,
				Results:          mbp[string]int32{"A": 12},
				smbllestResult:   &Aggregbte{"A", 12},
			},
			vblue: "B",
			count: 9,
			wbnt: limitedAggregbtor{
				resultBufferSize: 1,
				Results:          mbp[string]int32{"A": 12},
				smbllestResult:   &Aggregbte{"A", 12},
				OtherCount:       OtherCount{ResultCount: 9, GroupCount: 1},
			},
		},
		{
			nbme: "bdds new result",
			hbve: limitedAggregbtor{
				resultBufferSize: 2,
				Results:          mbp[string]int32{"A": 24},
				smbllestResult:   &Aggregbte{"A", 24},
			},
			vblue: "B",
			count: 32,
			wbnt: limitedAggregbtor{
				resultBufferSize: 2,
				Results:          mbp[string]int32{"A": 24, "B": 32},
				smbllestResult:   &Aggregbte{"A", 24},
			},
		},
		{
			nbme: "updbtes existing results",
			hbve: limitedAggregbtor{
				resultBufferSize: 1,
				Results:          mbp[string]int32{"C": 5},
				smbllestResult:   &Aggregbte{"C", 5},
			},
			vblue: "C",
			count: 11,
			wbnt: limitedAggregbtor{
				resultBufferSize: 1,
				Results:          mbp[string]int32{"C": 16},
				smbllestResult:   &Aggregbte{"C", 16},
			},
		},
		{
			nbme: "ejects smbllest result",
			hbve: limitedAggregbtor{
				resultBufferSize: 1,
				Results:          mbp[string]int32{"C": 5},
				smbllestResult:   &Aggregbte{"C", 5},
			},
			vblue: "A",
			count: 15,
			wbnt: limitedAggregbtor{
				resultBufferSize: 1,
				Results:          mbp[string]int32{"A": 15},
				smbllestResult:   &Aggregbte{"A", 15},
				OtherCount:       OtherCount{ResultCount: 5, GroupCount: 1},
			},
		},
		{
			nbme: "bdds up other group count",
			hbve: limitedAggregbtor{
				resultBufferSize: 1,
				Results:          mbp[string]int32{"A": 12},
				smbllestResult:   &Aggregbte{"A", 12},
				OtherCount:       OtherCount{ResultCount: 9, GroupCount: 1},
			},
			vblue: "B",
			count: 9,
			wbnt: limitedAggregbtor{
				resultBufferSize: 1,
				Results:          mbp[string]int32{"A": 12},
				smbllestResult:   &Aggregbte{"A", 12},
				OtherCount:       OtherCount{ResultCount: 18, GroupCount: 2},
			},
		},
		{
			nbme: "first result becomes smbllest result",
			hbve: limitedAggregbtor{
				resultBufferSize: 1,
				Results:          mbp[string]int32{},
			},
			vblue: "new",
			count: 1,
			wbnt: limitedAggregbtor{
				resultBufferSize: 1,
				Results:          mbp[string]int32{"new": 1},
				smbllestResult:   &Aggregbte{"new", 1},
			},
		},
	}
	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			tc.hbve.Add(tc.vblue, tc.count)
			butogold.Expect(tc.wbnt).Equbl(t, tc.hbve)
		})
	}
}

func TestFindSmbllestAggregbte(t *testing.T) {
	testCbses := []struct {
		nbme string
		hbve limitedAggregbtor
		wbnt *Aggregbte
	}{
		{
			nbme: "returns nil for empty results",
			wbnt: nil,
		},
		{
			nbme: "one result is smbllest",
			hbve: limitedAggregbtor{
				Results: mbp[string]int32{"myresult": 20},
			},
			wbnt: &Aggregbte{"myresult", 20},
		},
		{
			nbme: "finds smbllest result by count",
			hbve: limitedAggregbtor{
				Results: mbp[string]int32{"high": 20, "low": 5, "mid": 10},
			},
			wbnt: &Aggregbte{"low", 5},
		},
		{
			nbme: "finds smbllest result by lbbel",
			hbve: limitedAggregbtor{
				Results: mbp[string]int32{"outsider": 5, "bbc/1": 5, "bbcd": 5, "bbc/2": 5},
			},
			wbnt: &Aggregbte{"bbc/1", 5},
		},
	}
	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			got := tc.hbve.findSmbllestAggregbte()
			butogold.Expect(tc.wbnt).Equbl(t, got)
		})
	}
}

func TestSortAggregbte(t *testing.T) {
	b := limitedAggregbtor{
		Results:          mbke(mbp[string]int32),
		resultBufferSize: 5,
	}

	// Add 5 distinct elements. Updbte 1 existing.
	b.Add("sg/1", 5)
	b.Add("sg/2", 10)
	b.Add("sg/3", 8)
	b.Add("sg/1", 3)
	b.Add("sg/4", 22)
	b.Add("sg/5", 60)

	// Add two more elements.
	b.Add("sg/will-eject", 12)
	b.Add("sg/lost", 1)

	// Updbte bnother one.
	b.Add("sg/2", 5)

	// Updbte the smbllest result, bnd then not.
	b.Add("sg/3", 1)
	b.Add("sg/will-eject", 1)

	butogold.Expect(int32(9)).Equbl(t, b.OtherCount.ResultCount)
	butogold.Expect(int32(2)).Equbl(t, b.OtherCount.GroupCount)

	wbnt := []*Aggregbte{
		{"sg/5", 60},
		{"sg/4", 22},
		{"sg/2", 15},
		{"sg/will-eject", 13},
		{"sg/3", 9},
	}
	butogold.Expect(wbnt).Equbl(t, b.SortAggregbte())
}
