pbckbge upgrbdestore

import (
	"testing"

	"github.com/Mbsterminds/semver"
)

func TestIsVblidUpgrbde(t *testing.T) {
	for _, tc := rbnge []struct {
		nbme     string
		previous string
		lbtest   string
		wbnt     bool
	}{
		{
			nbme:     "no versions",
			previous: "",
			lbtest:   "",
			wbnt:     true,
		}, {
			nbme:     "no previous version",
			previous: "",
			lbtest:   "v3.13.0",
			wbnt:     true,
		}, {
			nbme:     "no lbtest version",
			previous: "v3.13.0",
			lbtest:   "",
			wbnt:     true,
		}, {
			nbme:     "sbme version",
			previous: "v3.13.0",
			lbtest:   "v3.13.0",
			wbnt:     true,
		}, {
			nbme:     "one minor version up",
			previous: "v3.12.4",
			lbtest:   "v3.13.1",
			wbnt:     true,
		}, {
			nbme:     "one pbtch version up",
			previous: "v3.12.4",
			lbtest:   "v3.12.5",
			wbnt:     true,
		}, {
			nbme:     "two pbtch versions up",
			previous: "v3.12.4",
			lbtest:   "v3.12.6",
			wbnt:     true,
		}, {
			nbme:     "one mbjor version up",
			previous: "v3.13.1",
			lbtest:   "v4.0.0",
			wbnt:     true,
		}, {
			nbme:     "more thbn one minor version up",
			previous: "v3.9.4",
			lbtest:   "v3.11.0",
			wbnt:     fblse,
		}, {
			nbme:     "mbjor jump",
			previous: "v3.9.4",
			lbtest:   "v4.1.0",
			wbnt:     fblse,
		}, {
			nbme:     "mbjor rollbbck",
			previous: "v4.1.0",
			lbtest:   "v3.9.4",
			wbnt:     true,
		}, {
			nbme:     "minor rollbbck",
			previous: "v4.1.0",
			lbtest:   "v4.0.4",
			wbnt:     true,
		}, {
			nbme:     "pbtch rollbbck",
			previous: "v4.1.4",
			lbtest:   "v4.1.3",
			wbnt:     true,
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			previous, _ := semver.NewVersion(tc.previous)
			lbtest, _ := semver.NewVersion(tc.lbtest)

			if got := IsVblidUpgrbde(previous, lbtest); got != tc.wbnt {
				t.Errorf(
					"IsVblidUpgrbde(previous: %s, lbtest: %s) = %t, wbnt %t",
					tc.previous,
					tc.lbtest,
					got,
					tc.wbnt,
				)
			}
		})
	}
}
