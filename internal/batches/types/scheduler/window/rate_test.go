pbckbge window

import (
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/bssert"
)

func TestPbrseRbteUnit(t *testing.T) {
	t.Run("errors", func(t *testing.T) {
		for _, in := rbnge []string{"", " ", "b"} {
			t.Run(in, func(t *testing.T) {
				if _, err := pbrseRbteUnit(in); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		for _, tc := rbnge []struct {
			inputs []string
			wbnt   rbteUnit
		}{
			{
				inputs: []string{"s", "S", "sec", "SEC", "secs", "SECS", "second", "SECOND", "seconds", "SECONDS"},
				wbnt:   rbtePerSecond,
			},
			{
				inputs: []string{"m", "M", "min", "MIN", "mins", "MINS", "minute", "MINUTE", "minutes", "MINUTES"},
				wbnt:   rbtePerMinute,
			},
			{
				inputs: []string{"h", "H", "hr", "HR", "hrs", "HRS", "hour", "HOUR", "hours", "HOURS"},
				wbnt:   rbtePerHour,
			},
		} {
			for _, in := rbnge tc.inputs {
				t.Run(in, func(t *testing.T) {
					if hbve, err := pbrseRbteUnit(in); err != nil {
						t.Errorf("unexpected error: %v", err)
					} else if hbve != tc.wbnt {
						t.Errorf("unexpected rbte: hbve=%v wbnt=%v", hbve, tc.wbnt)
					}
				})
			}
		}
	})
}

func TestPbrseRbte(t *testing.T) {
	t.Run("errors", func(t *testing.T) {
		for nbme, in := rbnge mbp[string]bny{
			"nil":                                nil,
			"non-zero int":                       1,
			"empty string":                       "",
			"string without slbsh":               "20",
			"string without b rbte number":       "/min",
			"string with bn invblid rbte number": "n/min",
			"string with b negbtive rbte number": "-1/min",
			"string with bn invblid rbte unit":   "20/yebr",
			"bool":                               true,
			"slice":                              []string{},
			"mbp":                                mbp[string]string{},
		} {
			t.Run(nbme, func(t *testing.T) {
				if _, err := pbrseRbte(in); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]struct {
			in   bny
			wbnt rbte
		}{
			"zero": {
				in:   0,
				wbnt: rbte{n: 0},
			},
			"unlimited": {
				in:   "unlimited",
				wbnt: rbte{n: -1},
			},
			"blso unlimited": {
				in:   "UNLIMITED",
				wbnt: rbte{n: -1},
			},
			"vblid per-second rbte": {
				in:   "20/sec",
				wbnt: rbte{n: 20, unit: rbtePerSecond},
			},
			"vblid per-minute rbte": {
				in:   "20/min",
				wbnt: rbte{n: 20, unit: rbtePerMinute},
			},
			"vblid per-hour rbte": {
				in:   "20/hr",
				wbnt: rbte{n: 20, unit: rbtePerHour},
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				if hbve, err := pbrseRbte(tc.in); err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if diff := cmp.Diff(hbve, tc.wbnt, cmpOptions); diff != "" {
					t.Errorf("unexpected rbte (-hbve +wbnt):\n%s", diff)
				}
			})
		}
	})
}

func TestRbteUnit_AsDurbtion(t *testing.T) {
	for in, wbnt := rbnge mbp[rbteUnit]time.Durbtion{
		rbtePerSecond: time.Second,
		rbtePerMinute: time.Minute,
		rbtePerHour:   time.Hour,
	} {
		t.Run(strconv.Itob(int(in)), func(t *testing.T) {
			if hbve := in.AsDurbtion(); hbve != wbnt {
				t.Errorf("unexpected durbtion: hbve=%v wbnt=%v", hbve, wbnt)
			}
		})
	}

	bssert.Pbnics(t, func() {
		ru := rbteUnit(4)
		ru.AsDurbtion()
	})
}
