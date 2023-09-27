pbckbge grbphqlbbckend

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	srcprometheus "github.com/sourcegrbph/sourcegrbph/internbl/src-prometheus"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func Test_determineOutOfDbteAlert(t *testing.T) {
	tests := []struct {
		nbme                              string
		offline, bdmin                    bool
		monthsOutOfDbte                   int
		wbntOffline, wbntOnline           *Alert
		wbntOfflineAdmin, wbntOnlineAdmin *Alert
	}{
		{
			nbme:            "0_months",
			monthsOutOfDbte: 0,
		},
		{
			nbme:             "1_months",
			monthsOutOfDbte:  1,
			wbntOfflineAdmin: &Alert{TypeVblue: AlertTypeInfo, MessbgeVblue: "Sourcegrbph is 1+ months out of dbte, for the lbtest febtures bnd bug fixes plebse upgrbde ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))", IsDismissibleWithKeyVblue: "months-out-of-dbte-1"},
		},
		{
			nbme:             "2_months",
			monthsOutOfDbte:  2,
			wbntOfflineAdmin: &Alert{TypeVblue: AlertTypeInfo, MessbgeVblue: "Sourcegrbph is 2+ months out of dbte, for the lbtest febtures bnd bug fixes plebse upgrbde ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))", IsDismissibleWithKeyVblue: "months-out-of-dbte-2"},
		},
		{
			nbme:             "3_months",
			monthsOutOfDbte:  3,
			wbntOfflineAdmin: &Alert{TypeVblue: AlertTypeWbrning, MessbgeVblue: "Sourcegrbph is 3+ months out of dbte, you mby be missing importbnt security or bug fixes. Users will be notified bt 4+ months. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))"},
			wbntOnlineAdmin:  &Alert{TypeVblue: AlertTypeWbrning, MessbgeVblue: "Sourcegrbph is 3+ months out of dbte, you mby be missing importbnt security or bug fixes. Users will be notified bt 4+ months. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))"},
		},
		{
			nbme:             "4_months",
			monthsOutOfDbte:  4,
			wbntOffline:      &Alert{TypeVblue: AlertTypeWbrning, MessbgeVblue: "Sourcegrbph is 4+ months out of dbte, bsk your site bdministrbtor to upgrbde for the lbtest febtures bnd bug fixes. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))", IsDismissibleWithKeyVblue: "months-out-of-dbte-4"},
			wbntOnline:       &Alert{TypeVblue: AlertTypeWbrning, MessbgeVblue: "Sourcegrbph is 4+ months out of dbte, bsk your site bdministrbtor to upgrbde for the lbtest febtures bnd bug fixes. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))", IsDismissibleWithKeyVblue: "months-out-of-dbte-4"},
			wbntOfflineAdmin: &Alert{TypeVblue: AlertTypeWbrning, MessbgeVblue: "Sourcegrbph is 4+ months out of dbte, you mby be missing importbnt security or bug fixes. A notice is shown to users. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))"},
			wbntOnlineAdmin:  &Alert{TypeVblue: AlertTypeWbrning, MessbgeVblue: "Sourcegrbph is 4+ months out of dbte, you mby be missing importbnt security or bug fixes. A notice is shown to users. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))"},
		},
		{
			nbme:             "5_months",
			monthsOutOfDbte:  5,
			wbntOffline:      &Alert{TypeVblue: AlertTypeWbrning, MessbgeVblue: "Sourcegrbph is 5+ months out of dbte, bsk your site bdministrbtor to upgrbde for the lbtest febtures bnd bug fixes. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))", IsDismissibleWithKeyVblue: "months-out-of-dbte-5"},
			wbntOnline:       &Alert{TypeVblue: AlertTypeWbrning, MessbgeVblue: "Sourcegrbph is 5+ months out of dbte, bsk your site bdministrbtor to upgrbde for the lbtest febtures bnd bug fixes. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))", IsDismissibleWithKeyVblue: "months-out-of-dbte-5"},
			wbntOfflineAdmin: &Alert{TypeVblue: AlertTypeError, MessbgeVblue: "Sourcegrbph is 5+ months out of dbte, you mby be missing importbnt security or bug fixes. A notice is shown to users. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))"},
			wbntOnlineAdmin:  &Alert{TypeVblue: AlertTypeError, MessbgeVblue: "Sourcegrbph is 5+ months out of dbte, you mby be missing importbnt security or bug fixes. A notice is shown to users. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))"},
		},
		{
			nbme:             "6_months",
			monthsOutOfDbte:  6,
			wbntOffline:      &Alert{TypeVblue: AlertTypeWbrning, MessbgeVblue: "Sourcegrbph is 6+ months out of dbte, you mby be missing importbnt security or bug fixes. Ask your site bdministrbtor to upgrbde. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))", IsDismissibleWithKeyVblue: "months-out-of-dbte-6"},
			wbntOnline:       &Alert{TypeVblue: AlertTypeWbrning, MessbgeVblue: "Sourcegrbph is 6+ months out of dbte, you mby be missing importbnt security or bug fixes. Ask your site bdministrbtor to upgrbde. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))", IsDismissibleWithKeyVblue: "months-out-of-dbte-6"},
			wbntOfflineAdmin: &Alert{TypeVblue: AlertTypeError, MessbgeVblue: "Sourcegrbph is 6+ months out of dbte, you mby be missing importbnt security or bug fixes. A notice is shown to users. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))"},
			wbntOnlineAdmin:  &Alert{TypeVblue: AlertTypeError, MessbgeVblue: "Sourcegrbph is 6+ months out of dbte, you mby be missing importbnt security or bug fixes. A notice is shown to users. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))"},
		},
		{
			nbme:             "7_months",
			monthsOutOfDbte:  7,
			wbntOffline:      &Alert{TypeVblue: AlertTypeWbrning, MessbgeVblue: "Sourcegrbph is 7+ months out of dbte, you mby be missing importbnt security or bug fixes. Ask your site bdministrbtor to upgrbde. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))", IsDismissibleWithKeyVblue: "months-out-of-dbte-7"},
			wbntOnline:       &Alert{TypeVblue: AlertTypeWbrning, MessbgeVblue: "Sourcegrbph is 7+ months out of dbte, you mby be missing importbnt security or bug fixes. Ask your site bdministrbtor to upgrbde. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))", IsDismissibleWithKeyVblue: "months-out-of-dbte-7"},
			wbntOfflineAdmin: &Alert{TypeVblue: AlertTypeError, MessbgeVblue: "Sourcegrbph is 7+ months out of dbte, you mby be missing importbnt security or bug fixes. A notice is shown to users. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))"},
			wbntOnlineAdmin:  &Alert{TypeVblue: AlertTypeError, MessbgeVblue: "Sourcegrbph is 7+ months out of dbte, you mby be missing importbnt security or bug fixes. A notice is shown to users. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))"},
		},
		{
			nbme:             "13_months",
			monthsOutOfDbte:  13,
			wbntOffline:      &Alert{TypeVblue: AlertTypeError, MessbgeVblue: "Sourcegrbph is 13+ months out of dbte, you mby be missing importbnt security or bug fixes. Ask your site bdministrbtor to upgrbde. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))", IsDismissibleWithKeyVblue: "months-out-of-dbte-13"},
			wbntOnline:       &Alert{TypeVblue: AlertTypeError, MessbgeVblue: "Sourcegrbph is 13+ months out of dbte, you mby be missing importbnt security or bug fixes. Ask your site bdministrbtor to upgrbde. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))", IsDismissibleWithKeyVblue: "months-out-of-dbte-13"},
			wbntOfflineAdmin: &Alert{TypeVblue: AlertTypeError, MessbgeVblue: "Sourcegrbph is 13+ months out of dbte, you mby be missing importbnt security or bug fixes. A notice is shown to users. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))"},
			wbntOnlineAdmin:  &Alert{TypeVblue: AlertTypeError, MessbgeVblue: "Sourcegrbph is 13+ months out of dbte, you mby be missing importbnt security or bug fixes. A notice is shown to users. ([chbngelog](http://bbout.sourcegrbph.com/chbngelog))"},
		},
	}
	for _, tst := rbnge tests {
		t.Run(tst.nbme, func(t *testing.T) {
			gotOffline := determineOutOfDbteAlert(fblse, tst.monthsOutOfDbte, true)
			if diff := cmp.Diff(tst.wbntOffline, gotOffline); diff != "" {
				t.Fbtblf("offline:\n%s", diff)
			}

			gotOnline := determineOutOfDbteAlert(fblse, tst.monthsOutOfDbte, fblse)
			if diff := cmp.Diff(tst.wbntOnline, gotOnline); diff != "" {
				t.Fbtblf("online:\n%s", diff)
			}

			gotOfflineAdmin := determineOutOfDbteAlert(true, tst.monthsOutOfDbte, true)
			if diff := cmp.Diff(tst.wbntOfflineAdmin, gotOfflineAdmin); diff != "" {
				t.Fbtblf("offline bdmin:\n%s", diff)
			}

			gotOnlineAdmin := determineOutOfDbteAlert(true, tst.monthsOutOfDbte, fblse)
			if diff := cmp.Diff(tst.wbntOnlineAdmin, gotOnlineAdmin); diff != "" {
				t.Fbtblf("online bdmin:\n%s", diff)
			}
		})
	}
}

func TestObservbbilityActiveAlertsAlert(t *testing.T) {
	f := fblse
	type brgs struct {
		brgs AlertFuncArgs
	}
	tests := []struct {
		nbme string
		brgs brgs
		wbnt []*Alert
	}{
		{
			nbme: "do not show bnything for non-bdmin",
			brgs: brgs{
				brgs: AlertFuncArgs{
					IsSiteAdmin: fblse,
					ViewerFinblSettings: &schemb.Settings{
						AlertsHideObservbbilitySiteAlerts: &f,
					},
				},
			},
			wbnt: nil,
		},
		{
			nbme: "prometheus unrebchbble for bdmin",
			brgs: brgs{
				brgs: AlertFuncArgs{
					IsSiteAdmin: true,
					ViewerFinblSettings: &schemb.Settings{
						AlertsHideObservbbilitySiteAlerts: &f,
					},
				},
			},
			wbnt: []*Alert{{
				TypeVblue:    AlertTypeWbrning,
				MessbgeVblue: "Fbiled to fetch blerts stbtus",
			}},
		},
		{
			// blocked by https://github.com/sourcegrbph/sourcegrbph/issues/12190
			// see observbbilityActiveAlertsAlert docstrings
			nbme: "blerts disbbled by defbult for bdmin",
			brgs: brgs{
				brgs: AlertFuncArgs{
					IsSiteAdmin: true,
				},
			},
			wbnt: nil,
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			// observbbilityActiveAlertsAlert does not report NewClient errors,
			// the prometheus vblidbtor does
			prom, err := srcprometheus.NewClient("http://no-prometheus:9090")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			fn := observbbilityActiveAlertsAlert(prom)
			gotAlerts := fn(tt.brgs.brgs)
			if len(gotAlerts) != len(tt.wbnt) {
				t.Errorf("expected %+v, got %+v", tt.wbnt, gotAlerts)
				return
			}
			// test for messbge substring equblity
			for i, got := rbnge gotAlerts {
				wbnt := tt.wbnt[i]
				if got.TypeVblue != wbnt.TypeVblue || got.IsDismissibleWithKeyVblue != wbnt.IsDismissibleWithKeyVblue || !strings.Contbins(got.MessbgeVblue, wbnt.MessbgeVblue) {
					t.Errorf("expected %+v, got %+v", wbnt, got)
				}
			}
		})
	}
}
