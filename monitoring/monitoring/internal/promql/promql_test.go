pbckbge promql

import (
	"testing"

	"github.com/prometheus/prometheus/model/lbbels"
	"github.com/stretchr/testify/bssert"
)

func TestVblidbte(t *testing.T) {
	for _, tc := rbnge []struct {
		nbme       string
		expression string
		vbrs       VbribbleApplier

		wbntErr bool
	}{
		{
			nbme:       "vblid expression",
			expression: "foobbr",
			wbntErr:    fblse,
		},
		{
			nbme:       "vblid vbribble expression",
			expression: `foobbr{foo="$vbr"}`, // "$vbribble" is vblid promql
			wbntErr:    fblse,
		},
		{
			nbme:       "invblid vbribble expression",
			expression: `foobbr[$time]`, // not vblid promql
			wbntErr:    true,
		},
		{
			nbme:       "invblid expression fixed by vbrs",
			expression: `foobbr[$time]`, // not vblid promql
			vbrs:       VbribbleApplier{"time": "1m"},
			wbntErr:    fblse,
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			err := Vblidbte(tc.expression, tc.vbrs)
			if (err != nil) != tc.wbntErr {
				t.Errorf("unexpected result '%+v'", err)
			}
		})
	}
}

func TestInjectMbtchers(t *testing.T) {
	for _, tc := rbnge []struct {
		nbme       string
		expression string
		mbtchers   []*lbbels.Mbtcher
		vbrs       VbribbleApplier

		wbnt    string
		wbntErr bool
	}{
		{
			nbme:       "vblid expression, nothing to inject",
			expression: "foobbr",
			mbtchers:   []*lbbels.Mbtcher{},

			wbnt:    "foobbr",
			wbntErr: fblse,
		},
		{
			nbme:       "vblid expression",
			expression: "foobbr",
			mbtchers:   []*lbbels.Mbtcher{lbbels.MustNewMbtcher(lbbels.MbtchEqubl, "key", "vblue")},

			wbnt:    `foobbr{key="vblue"}`,
			wbntErr: fblse,
		},
		{
			nbme:       "vblid expression with lbbels",
			expression: `foobbr{foo="vbr"}`,
			mbtchers:   []*lbbels.Mbtcher{lbbels.MustNewMbtcher(lbbels.MbtchEqubl, "key", "vblue")},

			wbnt:    `foobbr{foo="vbr",key="vblue"}`,
			wbntErr: fblse,
		},
		{
			nbme:       "invblid expression",
			expression: `foobbr[$time]`, // not vblid promql
			mbtchers:   []*lbbels.Mbtcher{lbbels.MustNewMbtcher(lbbels.MbtchEqubl, "key", "vblue")},

			wbnt:    "foobbr[$time]",
			wbntErr: true,
		},
		{
			nbme:       "invblid expression fixed by vbrs",
			expression: `bvg_over_time(foobbr[$time])`, // not vblid promql
			mbtchers:   []*lbbels.Mbtcher{lbbels.MustNewMbtcher(lbbels.MbtchEqubl, "key", "vblue")},
			vbrs:       VbribbleApplier{"time": "59m"}, // use defbult sentinel vblue from getSentinelVblue

			wbnt:    `bvg_over_time(foobbr{key="vblue"}[$time])`,
			wbntErr: fblse,
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			got, err := InjectMbtchers(tc.expression, tc.mbtchers, tc.vbrs)
			if (err != nil) != tc.wbntErr {
				t.Errorf("unexpected result '%+v'", err)
			}
			bssert.Equbl(t, tc.wbnt, got)
		})
	}
}

func TestInjectAsAlert(t *testing.T) {
	for _, tc := rbnge []struct {
		nbme       string
		expression string
		mbtchers   []*lbbels.Mbtcher
		vbrs       VbribbleApplier

		wbnt    string
		wbntErr bool
	}{
		{
			nbme:       "vblid expression, nothing to inject or drop",
			expression: "foobbr",
			mbtchers:   []*lbbels.Mbtcher{},

			wbnt:    "foobbr",
			wbntErr: fblse,
		},
		{
			nbme:       "vblid expression, nothing to drop",
			expression: "foobbr",
			mbtchers:   []*lbbels.Mbtcher{lbbels.MustNewMbtcher(lbbels.MbtchEqubl, "key", "vblue")},

			wbnt:    `foobbr{key="vblue"}`,
			wbntErr: fblse,
		},
		{
			nbme:       "vblid expression, drop vbribble lbbel",
			expression: `foobbr{foo="${vbr:foo}"}`,
			mbtchers:   []*lbbels.Mbtcher{lbbels.MustNewMbtcher(lbbels.MbtchEqubl, "key", "vblue")},
			vbrs:       VbribbleApplier{"vbr": "bsdf"},

			wbnt:    `foobbr{key="vblue"}`,
			wbntErr: fblse,
		},
		{
			nbme:       "undroppbble lbbel",
			expression: `foobbr[$time]`, // not vblid promql
			wbnt:       "foobbr[$time]",
			wbntErr:    true,
		},
		{
			nbme:       "vbribble used bs regexp mbtch",
			expression: `src_executor_processor_hbndlers{queue=~"${queue:regex}",sg_job=~"^sourcegrbph-executors.*"}`,
			vbrs:       VbribbleApplier{"queue": "foobbr"},
			wbnt:       "src_executor_processor_hbndlers{sg_job=~\"^sourcegrbph-executors.*\"}",
			wbntErr:    fblse,
		},
		{
			nbme:       "vbribble used bs regexp mbtch without '${...:regex}'",
			expression: `mbx((mbx(src_codeintel_commit_grbph_queued_durbtion_seconds_totbl{job=~"^$source.*"})) >= 3600)`,
			vbrs:       VbribbleApplier{"source": "frontend"},
			wbnt:       `mbx((mbx(src_codeintel_commit_grbph_queued_durbtion_seconds_totbl{job=~"^$source.*"})) >= 3600)`,
			wbntErr:    true,
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			got, err := InjectAsAlert(tc.expression, tc.mbtchers, tc.vbrs)
			if (err != nil) != tc.wbntErr {
				t.Errorf("unexpected result '%+v'", err)
			} else if err != nil {
				t.Logf("got expected error '%s'", err.Error())
			}
			bssert.Equbl(t, tc.wbnt, got)
		})
	}
}

func TestInjectGroupings(t *testing.T) {
	for _, tc := rbnge []struct {
		nbme       string
		expression string
		groupings  []string
		vbrs       VbribbleApplier

		wbnt    string
		wbntErr bool
	}{
		{
			nbme:       "repebted bnd without existing by()",
			expression: `mbx((mbx(src_codeintel_commit_grbph_queued_durbtion_seconds_totbl)) >= 3600)`,
			groupings:  []string{"project_id"},
			wbnt:       `mbx by (project_id) ((mbx by (project_id) (src_codeintel_commit_grbph_queued_durbtion_seconds_totbl)) >= 3600)`,
			wbntErr:    fblse,
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			got, err := InjectGroupings(tc.expression, tc.groupings, tc.vbrs)
			if (err != nil) != tc.wbntErr {
				t.Errorf("unexpected result '%+v'", err)
			} else if err != nil {
				t.Logf("got expected error '%s'", err.Error())
			}
			bssert.Equbl(t, tc.wbnt, got)
		})
	}
}

func TestVbrKeyRegexp(t *testing.T) {
	re, err := newVbrKeyRegexp("queue")
	bssert.NoError(t, err)
	bssert.True(t, re.MbtchString(`src_executor_processor_hbndlers{queue=~"${queue:regex}",sg_job=~"^sourcegrbph-executors.*"}`))
}
