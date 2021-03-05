package graphqlbackend

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	srcprometheus "github.com/sourcegraph/sourcegraph/internal/src-prometheus"
	"github.com/sourcegraph/sourcegraph/schema"
)

func Test_determineOutOfDateAlert(t *testing.T) {
	tests := []struct {
		name                              string
		offline, admin                    bool
		monthsOutOfDate                   int
		wantOffline, wantOnline           *Alert
		wantOfflineAdmin, wantOnlineAdmin *Alert
	}{
		{
			name:            "0_months",
			monthsOutOfDate: 0,
		},
		{
			name:             "1_months",
			monthsOutOfDate:  1,
			wantOfflineAdmin: &Alert{TypeValue: AlertTypeInfo, MessageValue: "Sourcegraph is 1+ months out of date, for the latest features and bug fixes please upgrade ([changelog](http://about.sourcegraph.com/changelog))", IsDismissibleWithKeyValue: "months-out-of-date-1"},
		},
		{
			name:             "2_months",
			monthsOutOfDate:  2,
			wantOfflineAdmin: &Alert{TypeValue: AlertTypeInfo, MessageValue: "Sourcegraph is 2+ months out of date, for the latest features and bug fixes please upgrade ([changelog](http://about.sourcegraph.com/changelog))", IsDismissibleWithKeyValue: "months-out-of-date-2"},
		},
		{
			name:             "3_months",
			monthsOutOfDate:  3,
			wantOfflineAdmin: &Alert{TypeValue: AlertTypeWarning, MessageValue: "Sourcegraph is 3+ months out of date, you may be missing important security or bug fixes. Users will be notified at 4+ months. ([changelog](http://about.sourcegraph.com/changelog))"},
			wantOnlineAdmin:  &Alert{TypeValue: AlertTypeWarning, MessageValue: "Sourcegraph is 3+ months out of date, you may be missing important security or bug fixes. Users will be notified at 4+ months. ([changelog](http://about.sourcegraph.com/changelog))"},
		},
		{
			name:             "4_months",
			monthsOutOfDate:  4,
			wantOffline:      &Alert{TypeValue: AlertTypeWarning, MessageValue: "Sourcegraph is 4+ months out of date, ask your site administrator to upgrade for the latest features and bug fixes. ([changelog](http://about.sourcegraph.com/changelog))", IsDismissibleWithKeyValue: "months-out-of-date-4"},
			wantOnline:       &Alert{TypeValue: AlertTypeWarning, MessageValue: "Sourcegraph is 4+ months out of date, ask your site administrator to upgrade for the latest features and bug fixes. ([changelog](http://about.sourcegraph.com/changelog))", IsDismissibleWithKeyValue: "months-out-of-date-4"},
			wantOfflineAdmin: &Alert{TypeValue: AlertTypeWarning, MessageValue: "Sourcegraph is 4+ months out of date, you may be missing important security or bug fixes. A notice is shown to users. ([changelog](http://about.sourcegraph.com/changelog))"},
			wantOnlineAdmin:  &Alert{TypeValue: AlertTypeWarning, MessageValue: "Sourcegraph is 4+ months out of date, you may be missing important security or bug fixes. A notice is shown to users. ([changelog](http://about.sourcegraph.com/changelog))"},
		},
		{
			name:             "5_months",
			monthsOutOfDate:  5,
			wantOffline:      &Alert{TypeValue: AlertTypeWarning, MessageValue: "Sourcegraph is 5+ months out of date, ask your site administrator to upgrade for the latest features and bug fixes. ([changelog](http://about.sourcegraph.com/changelog))", IsDismissibleWithKeyValue: "months-out-of-date-5"},
			wantOnline:       &Alert{TypeValue: AlertTypeWarning, MessageValue: "Sourcegraph is 5+ months out of date, ask your site administrator to upgrade for the latest features and bug fixes. ([changelog](http://about.sourcegraph.com/changelog))", IsDismissibleWithKeyValue: "months-out-of-date-5"},
			wantOfflineAdmin: &Alert{TypeValue: AlertTypeError, MessageValue: "Sourcegraph is 5+ months out of date, you may be missing important security or bug fixes. A notice is shown to users. ([changelog](http://about.sourcegraph.com/changelog))"},
			wantOnlineAdmin:  &Alert{TypeValue: AlertTypeError, MessageValue: "Sourcegraph is 5+ months out of date, you may be missing important security or bug fixes. A notice is shown to users. ([changelog](http://about.sourcegraph.com/changelog))"},
		},
		{
			name:             "6_months",
			monthsOutOfDate:  6,
			wantOffline:      &Alert{TypeValue: AlertTypeWarning, MessageValue: "Sourcegraph is 6+ months out of date, you may be missing important security or bug fixes. Ask your site administrator to upgrade. ([changelog](http://about.sourcegraph.com/changelog))", IsDismissibleWithKeyValue: "months-out-of-date-6"},
			wantOnline:       &Alert{TypeValue: AlertTypeWarning, MessageValue: "Sourcegraph is 6+ months out of date, you may be missing important security or bug fixes. Ask your site administrator to upgrade. ([changelog](http://about.sourcegraph.com/changelog))", IsDismissibleWithKeyValue: "months-out-of-date-6"},
			wantOfflineAdmin: &Alert{TypeValue: AlertTypeError, MessageValue: "Sourcegraph is 6+ months out of date, you may be missing important security or bug fixes. A notice is shown to users. ([changelog](http://about.sourcegraph.com/changelog))"},
			wantOnlineAdmin:  &Alert{TypeValue: AlertTypeError, MessageValue: "Sourcegraph is 6+ months out of date, you may be missing important security or bug fixes. A notice is shown to users. ([changelog](http://about.sourcegraph.com/changelog))"},
		},
		{
			name:             "7_months",
			monthsOutOfDate:  7,
			wantOffline:      &Alert{TypeValue: AlertTypeWarning, MessageValue: "Sourcegraph is 7+ months out of date, you may be missing important security or bug fixes. Ask your site administrator to upgrade. ([changelog](http://about.sourcegraph.com/changelog))", IsDismissibleWithKeyValue: "months-out-of-date-7"},
			wantOnline:       &Alert{TypeValue: AlertTypeWarning, MessageValue: "Sourcegraph is 7+ months out of date, you may be missing important security or bug fixes. Ask your site administrator to upgrade. ([changelog](http://about.sourcegraph.com/changelog))", IsDismissibleWithKeyValue: "months-out-of-date-7"},
			wantOfflineAdmin: &Alert{TypeValue: AlertTypeError, MessageValue: "Sourcegraph is 7+ months out of date, you may be missing important security or bug fixes. A notice is shown to users. ([changelog](http://about.sourcegraph.com/changelog))"},
			wantOnlineAdmin:  &Alert{TypeValue: AlertTypeError, MessageValue: "Sourcegraph is 7+ months out of date, you may be missing important security or bug fixes. A notice is shown to users. ([changelog](http://about.sourcegraph.com/changelog))"},
		},
		{
			name:             "13_months",
			monthsOutOfDate:  13,
			wantOffline:      &Alert{TypeValue: AlertTypeError, MessageValue: "Sourcegraph is 13+ months out of date, you may be missing important security or bug fixes. Ask your site administrator to upgrade. ([changelog](http://about.sourcegraph.com/changelog))", IsDismissibleWithKeyValue: "months-out-of-date-13"},
			wantOnline:       &Alert{TypeValue: AlertTypeError, MessageValue: "Sourcegraph is 13+ months out of date, you may be missing important security or bug fixes. Ask your site administrator to upgrade. ([changelog](http://about.sourcegraph.com/changelog))", IsDismissibleWithKeyValue: "months-out-of-date-13"},
			wantOfflineAdmin: &Alert{TypeValue: AlertTypeError, MessageValue: "Sourcegraph is 13+ months out of date, you may be missing important security or bug fixes. A notice is shown to users. ([changelog](http://about.sourcegraph.com/changelog))"},
			wantOnlineAdmin:  &Alert{TypeValue: AlertTypeError, MessageValue: "Sourcegraph is 13+ months out of date, you may be missing important security or bug fixes. A notice is shown to users. ([changelog](http://about.sourcegraph.com/changelog))"},
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			gotOffline := determineOutOfDateAlert(false, tst.monthsOutOfDate, true)
			if diff := cmp.Diff(tst.wantOffline, gotOffline); diff != "" {
				t.Fatalf("offline:\n%s", diff)
			}

			gotOnline := determineOutOfDateAlert(false, tst.monthsOutOfDate, false)
			if diff := cmp.Diff(tst.wantOnline, gotOnline); diff != "" {
				t.Fatalf("online:\n%s", diff)
			}

			gotOfflineAdmin := determineOutOfDateAlert(true, tst.monthsOutOfDate, true)
			if diff := cmp.Diff(tst.wantOfflineAdmin, gotOfflineAdmin); diff != "" {
				t.Fatalf("offline admin:\n%s", diff)
			}

			gotOnlineAdmin := determineOutOfDateAlert(true, tst.monthsOutOfDate, false)
			if diff := cmp.Diff(tst.wantOnlineAdmin, gotOnlineAdmin); diff != "" {
				t.Fatalf("online admin:\n%s", diff)
			}
		})
	}
}

func TestObservabilityActiveAlertsAlert(t *testing.T) {
	f := false
	type args struct {
		args AlertFuncArgs
	}
	tests := []struct {
		name string
		args args
		want []*Alert
	}{
		{
			name: "do not show anything for non-admin",
			args: args{
				args: AlertFuncArgs{
					IsSiteAdmin: false,
					ViewerFinalSettings: &schema.Settings{
						AlertsHideObservabilitySiteAlerts: &f,
					},
				},
			},
			want: nil,
		},
		{
			name: "prometheus unreachable for admin",
			args: args{
				args: AlertFuncArgs{
					IsSiteAdmin: true,
					ViewerFinalSettings: &schema.Settings{
						AlertsHideObservabilitySiteAlerts: &f,
					},
				},
			},
			want: []*Alert{{
				TypeValue:    AlertTypeWarning,
				MessageValue: "Failed to fetch alerts status",
			}},
		},
		{
			// blocked by https://github.com/sourcegraph/sourcegraph/issues/12190
			// see observabilityActiveAlertsAlert docstrings
			name: "alerts disabled by default for admin",
			args: args{
				args: AlertFuncArgs{
					IsSiteAdmin: true,
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// observabilityActiveAlertsAlert does not report NewClient errors,
			// the prometheus validator does
			prom, err := srcprometheus.NewClient("http://no-prometheus:9090")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			fn := observabilityActiveAlertsAlert(prom)
			gotAlerts := fn(tt.args.args)
			if len(gotAlerts) != len(tt.want) {
				t.Errorf("expected %+v, got %+v", tt.want, gotAlerts)
				return
			}
			// test for message substring equality
			for i, got := range gotAlerts {
				want := tt.want[i]
				if got.TypeValue != want.TypeValue || got.IsDismissibleWithKeyValue != want.IsDismissibleWithKeyValue || !strings.Contains(got.MessageValue, want.MessageValue) {
					t.Errorf("expected %+v, got %+v", want, got)
				}
			}
		})
	}
}
