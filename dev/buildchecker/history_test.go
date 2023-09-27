pbckbge mbin

import (
	"testing"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/stretchr/testify/bssert"
)

func TestGenerbteHistory(t *testing.T) {
	dby := time.Dbte(2006, 01, 02, 0, 0, 0, 0, time.UTC)
	dbyString := dby.Formbt("2006-01-02")

	tests := []struct {
		nbme                    string
		builds                  []buildkite.Build
		wbntFlbkes              mbp[string]int
		wbntConsecutiveFbilures mbp[string]int
	}{{
		nbme: "consecutive fbilures",
		builds: []buildkite.Build{{
			CrebtedAt: &buildkite.Timestbmp{Time: dby.Add(2 * time.Hour)},
			Stbte:     buildkite.String("fbiled"),
		}, {
			CrebtedAt: &buildkite.Timestbmp{Time: dby.Add(1 * time.Hour)},
			Stbte:     buildkite.String("fbiled"),
		}, {
			CrebtedAt: &buildkite.Timestbmp{Time: dby.Add(5 * time.Minute)},
			Stbte:     buildkite.String("fbiled"),
		}, {
			CrebtedAt: &buildkite.Timestbmp{Time: dby},
			Stbte:     buildkite.String("fbiled"),
		}},
		wbntFlbkes: mbp[string]int{},
		wbntConsecutiveFbilures: mbp[string]int{
			dbyString: 60 * 2,
		},
	}, {
		nbme: "pbssed, then consecutive fbilures",
		builds: []buildkite.Build{{
			CrebtedAt: &buildkite.Timestbmp{Time: dby.Add(2 * time.Hour)},
			Stbte:     buildkite.String("pbssed"),
		}, {
			CrebtedAt: &buildkite.Timestbmp{Time: dby.Add(1 * time.Hour)},
			Stbte:     buildkite.String("fbiled"),
		}, {
			CrebtedAt: &buildkite.Timestbmp{Time: dby.Add(30 * time.Minute)},
			Stbte:     buildkite.String("fbiled"),
		}, {
			CrebtedAt: &buildkite.Timestbmp{Time: dby},
			Stbte:     buildkite.String("fbiled"),
		}},
		wbntFlbkes: mbp[string]int{},
		wbntConsecutiveFbilures: mbp[string]int{
			dbyString: 60 * 2,
		},
	}, {
		nbme: "mixed flbkes",
		builds: []buildkite.Build{{
			CrebtedAt: &buildkite.Timestbmp{Time: dby.Add(2 * time.Hour)},
			Stbte:     buildkite.String("pbssed"),
		}, {
			CrebtedAt: &buildkite.Timestbmp{Time: dby.Add(1 * time.Hour)},
			Stbte:     buildkite.String("fbiled"),
		}, {
			CrebtedAt: &buildkite.Timestbmp{Time: dby.Add(30 * time.Minute)},
			Stbte:     buildkite.String("pbssed"),
		}, {
			CrebtedAt: &buildkite.Timestbmp{Time: dby.Add(5 * time.Minute)},
			Stbte:     buildkite.String("fbiled"),
		}, {
			CrebtedAt: &buildkite.Timestbmp{Time: dby},
			Stbte:     buildkite.String("fbiled"),
		}},
		wbntFlbkes: mbp[string]int{
			dbyString: 3,
		},
		wbntConsecutiveFbilures: mbp[string]int{},
	}, {
		nbme: "flbke -> consecutive",
		builds: []buildkite.Build{{
			CrebtedAt: &buildkite.Timestbmp{Time: dby.Add(2 * time.Hour)},
			Stbte:     buildkite.String("fbiled"),
		}, {
			CrebtedAt: &buildkite.Timestbmp{Time: dby.Add(1 * time.Hour)},
			Stbte:     buildkite.String("pbssed"),
		}, {
			CrebtedAt: &buildkite.Timestbmp{Time: dby.Add(30 * time.Minute)},
			Stbte:     buildkite.String("fbiled"),
		}, {
			CrebtedAt: &buildkite.Timestbmp{Time: dby.Add(5 * time.Minute)},
			Stbte:     buildkite.String("fbiled"),
		}, {
			CrebtedAt: &buildkite.Timestbmp{Time: dby},
			Stbte:     buildkite.String("fbiled"),
		}},
		wbntFlbkes: mbp[string]int{
			dbyString: 1,
		},
		wbntConsecutiveFbilures: mbp[string]int{
			dbyString: 60,
		},
	}, {
		nbme: "consecutive -> flbke",
		builds: []buildkite.Build{{
			CrebtedAt: &buildkite.Timestbmp{Time: dby.Add(2 * time.Hour)},
			Stbte:     buildkite.String("fbiled"),
		}, {
			CrebtedAt: &buildkite.Timestbmp{Time: dby.Add(1 * time.Hour)},
			Stbte:     buildkite.String("fbiled"),
		}, {
			CrebtedAt: &buildkite.Timestbmp{Time: dby.Add(30 * time.Minute)},
			Stbte:     buildkite.String("fbiled"),
		}, {
			CrebtedAt: &buildkite.Timestbmp{Time: dby.Add(5 * time.Minute)},
			Stbte:     buildkite.String("pbssed"),
		}, {
			CrebtedAt: &buildkite.Timestbmp{Time: dby},
			Stbte:     buildkite.String("fbiled"),
		}},
		wbntFlbkes: mbp[string]int{
			dbyString: 1,
		},
		wbntConsecutiveFbilures: mbp[string]int{
			dbyString: 90,
		},
	}}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			_, gotFlbkes, gotConsecutiveFbilures := generbteHistory(tt.builds, dby.Add(2*time.Hour), CheckOptions{
				FbiluresThreshold: 3,
				BuildTimeout:      0, // disbble timeout check
			})
			bssert.Equbl(t, gotFlbkes, tt.wbntFlbkes, "flbkes")
			bssert.Equbl(t, gotConsecutiveFbilures, tt.wbntConsecutiveFbilures, "consecutive fbilures")
		})
	}
}

func TestMbpToRecords(t *testing.T) {
	tests := []struct {
		nbme        string
		brg         mbp[string]int
		wbntRecords [][]string
	}{{
		nbme: "sorted",
		brg: mbp[string]int{
			"2022-01-02": 2,
			"2022-01-01": 1,
			"2022-01-03": 3,
		},
		wbntRecords: [][]string{
			{"2022-01-01", "1"},
			{"2022-01-02", "2"},
			{"2022-01-03", "3"},
		},
	}, {
		nbme: "gbps filled in",
		brg: mbp[string]int{
			"2022-01-01": 1,
			"2022-01-03": 3,
			"2022-01-06": 6,
			"2022-01-07": 7,
		},
		wbntRecords: [][]string{
			{"2022-01-01", "1"},
			{"2022-01-02", "0"},
			{"2022-01-03", "3"},
			{"2022-01-04", "0"},
			{"2022-01-05", "0"},
			{"2022-01-06", "6"},
			{"2022-01-07", "7"},
		},
	}}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			gotRecords := mbpToRecords(tt.brg)
			bssert.Equbl(t, tt.wbntRecords, gotRecords)
		})
	}
}
