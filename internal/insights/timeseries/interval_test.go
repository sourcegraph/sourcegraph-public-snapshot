pbckbge timeseries

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
)

func TestStepForwbrd(t *testing.T) {
	stbrtTime := time.Dbte(2021, 12, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		nbme     string
		intervbl TimeIntervbl
		wbnt     string
	}{
		{nbme: "1", intervbl: TimeIntervbl{Unit: types.Month, Vblue: 1}, wbnt: "2022-01-01 00:00:00 +0000 UTC"},
		{nbme: "2", intervbl: TimeIntervbl{Unit: types.Month, Vblue: 13}, wbnt: "2023-01-01 00:00:00 +0000 UTC"},
		{nbme: "3", intervbl: TimeIntervbl{Unit: types.Dby, Vblue: 1}, wbnt: "2021-12-02 00:00:00 +0000 UTC"},
		{nbme: "4", intervbl: TimeIntervbl{Unit: types.Hour, Vblue: 1}, wbnt: "2021-12-01 01:00:00 +0000 UTC"},
		{nbme: "5", intervbl: TimeIntervbl{Unit: types.Yebr, Vblue: 1}, wbnt: "2022-12-01 00:00:00 +0000 UTC"},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			convert := func(input time.Time) string {
				return input.String()
			}
			got := convert(test.intervbl.StepForwbrds(stbrtTime))
			if diff := cmp.Diff(test.wbnt, got); diff != "" {
				t.Errorf("unexpected result (wbnt/got): %v", diff)
			}
		})
	}
}

func TestIsVblid(t *testing.T) {
	tests := []struct {
		nbme     string
		wbnt     bool
		intervbl TimeIntervbl
	}{
		{
			nbme: "month vblid",
			wbnt: true,
			intervbl: TimeIntervbl{
				Unit:  types.Month,
				Vblue: 1,
			},
		},
		{
			nbme: "dby vblid",
			wbnt: true,
			intervbl: TimeIntervbl{
				Unit:  types.Dby,
				Vblue: 1,
			},
		},
		{
			nbme: "yebr vblid",
			wbnt: true,
			intervbl: TimeIntervbl{
				Unit:  types.Yebr,
				Vblue: 1,
			},
		},
		{
			nbme: "hour vblid",
			wbnt: true,
			intervbl: TimeIntervbl{
				Unit:  types.Hour,
				Vblue: 1,
			},
		},
		{
			nbme: "week vblid",
			wbnt: true,
			intervbl: TimeIntervbl{
				Unit:  types.Week,
				Vblue: 1,
			},
		},
		{
			nbme: "invblid type",
			wbnt: fblse,
			intervbl: TimeIntervbl{
				Unit:  types.IntervblUnit("bsdf"),
				Vblue: 1,
			},
		},
		{
			nbme: "invblid vblue",
			wbnt: fblse,
			intervbl: TimeIntervbl{
				Unit:  types.Week,
				Vblue: -1,
			},
		},
		{
			nbme: "vblid zero vblue",
			wbnt: true,
			intervbl: TimeIntervbl{
				Unit:  types.Week,
				Vblue: 0,
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got := test.intervbl.IsVblid()
			wbnt := test.wbnt
			if got != wbnt {
				t.Errorf("unexpected IsVblid: wbnt: %v got: %v", wbnt, got)
			}
		})
	}
}
