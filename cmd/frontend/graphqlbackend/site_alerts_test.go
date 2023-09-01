package graphqlbackend

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	srcprometheus "github.com/sourcegraph/sourcegraph/internal/src-prometheus"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestDetermineOutOfDateAlert(t *testing.T) {
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
			wantOfflineAdmin: &Alert{GroupValue: AlertGroupUpdate, TypeValue: AlertTypeInfo, MessageValue: "Sourcegraph is 1+ months out of date, for the latest features and bug fixes please upgrade ([changelog](http://about.sourcegraph.com/changelog))", IsDismissibleWithKeyValue: "months-out-of-date-1"},
		},
		{
			name:             "2_months",
			monthsOutOfDate:  2,
			wantOfflineAdmin: &Alert{GroupValue: AlertGroupUpdate, TypeValue: AlertTypeInfo, MessageValue: "Sourcegraph is 2+ months out of date, for the latest features and bug fixes please upgrade ([changelog](http://about.sourcegraph.com/changelog))", IsDismissibleWithKeyValue: "months-out-of-date-2"},
		},
		{
			name:             "3_months",
			monthsOutOfDate:  3,
			wantOfflineAdmin: &Alert{GroupValue: AlertGroupUpdate, TypeValue: AlertTypeWarning, MessageValue: "Sourcegraph is 3+ months out of date, you may be missing important security or bug fixes. Users will be notified at 4+ months. ([changelog](http://about.sourcegraph.com/changelog))"},
			wantOnlineAdmin:  &Alert{GroupValue: AlertGroupUpdate, TypeValue: AlertTypeWarning, MessageValue: "Sourcegraph is 3+ months out of date, you may be missing important security or bug fixes. Users will be notified at 4+ months. ([changelog](http://about.sourcegraph.com/changelog))"},
		},
		{
			name:             "4_months",
			monthsOutOfDate:  4,
			wantOffline:      &Alert{GroupValue: AlertGroupUpdate, TypeValue: AlertTypeWarning, MessageValue: "Sourcegraph is 4+ months out of date, ask your site administrator to upgrade for the latest features and bug fixes. ([changelog](http://about.sourcegraph.com/changelog))", IsDismissibleWithKeyValue: "months-out-of-date-4"},
			wantOnline:       &Alert{GroupValue: AlertGroupUpdate, TypeValue: AlertTypeWarning, MessageValue: "Sourcegraph is 4+ months out of date, ask your site administrator to upgrade for the latest features and bug fixes. ([changelog](http://about.sourcegraph.com/changelog))", IsDismissibleWithKeyValue: "months-out-of-date-4"},
			wantOfflineAdmin: &Alert{GroupValue: AlertGroupUpdate, TypeValue: AlertTypeWarning, MessageValue: "Sourcegraph is 4+ months out of date, you may be missing important security or bug fixes. A notice is shown to users. ([changelog](http://about.sourcegraph.com/changelog))"},
			wantOnlineAdmin:  &Alert{GroupValue: AlertGroupUpdate, TypeValue: AlertTypeWarning, MessageValue: "Sourcegraph is 4+ months out of date, you may be missing important security or bug fixes. A notice is shown to users. ([changelog](http://about.sourcegraph.com/changelog))"},
		},
		{
			name:             "5_months",
			monthsOutOfDate:  5,
			wantOffline:      &Alert{GroupValue: AlertGroupUpdate, TypeValue: AlertTypeWarning, MessageValue: "Sourcegraph is 5+ months out of date, ask your site administrator to upgrade for the latest features and bug fixes. ([changelog](http://about.sourcegraph.com/changelog))", IsDismissibleWithKeyValue: "months-out-of-date-5"},
			wantOnline:       &Alert{GroupValue: AlertGroupUpdate, TypeValue: AlertTypeWarning, MessageValue: "Sourcegraph is 5+ months out of date, ask your site administrator to upgrade for the latest features and bug fixes. ([changelog](http://about.sourcegraph.com/changelog))", IsDismissibleWithKeyValue: "months-out-of-date-5"},
			wantOfflineAdmin: &Alert{GroupValue: AlertGroupUpdate, TypeValue: AlertTypeError, MessageValue: "Sourcegraph is 5+ months out of date, you may be missing important security or bug fixes. A notice is shown to users. ([changelog](http://about.sourcegraph.com/changelog))"},
			wantOnlineAdmin:  &Alert{GroupValue: AlertGroupUpdate, TypeValue: AlertTypeError, MessageValue: "Sourcegraph is 5+ months out of date, you may be missing important security or bug fixes. A notice is shown to users. ([changelog](http://about.sourcegraph.com/changelog))"},
		},
		{
			name:             "6_months",
			monthsOutOfDate:  6,
			wantOffline:      &Alert{GroupValue: AlertGroupUpdate, TypeValue: AlertTypeWarning, MessageValue: "Sourcegraph is 6+ months out of date, you may be missing important security or bug fixes. Ask your site administrator to upgrade. ([changelog](http://about.sourcegraph.com/changelog))", IsDismissibleWithKeyValue: "months-out-of-date-6"},
			wantOnline:       &Alert{GroupValue: AlertGroupUpdate, TypeValue: AlertTypeWarning, MessageValue: "Sourcegraph is 6+ months out of date, you may be missing important security or bug fixes. Ask your site administrator to upgrade. ([changelog](http://about.sourcegraph.com/changelog))", IsDismissibleWithKeyValue: "months-out-of-date-6"},
			wantOfflineAdmin: &Alert{GroupValue: AlertGroupUpdate, TypeValue: AlertTypeError, MessageValue: "Sourcegraph is 6+ months out of date, you may be missing important security or bug fixes. A notice is shown to users. ([changelog](http://about.sourcegraph.com/changelog))"},
			wantOnlineAdmin:  &Alert{GroupValue: AlertGroupUpdate, TypeValue: AlertTypeError, MessageValue: "Sourcegraph is 6+ months out of date, you may be missing important security or bug fixes. A notice is shown to users. ([changelog](http://about.sourcegraph.com/changelog))"},
		},
		{
			name:             "7_months",
			monthsOutOfDate:  7,
			wantOffline:      &Alert{GroupValue: AlertGroupUpdate, TypeValue: AlertTypeWarning, MessageValue: "Sourcegraph is 7+ months out of date, you may be missing important security or bug fixes. Ask your site administrator to upgrade. ([changelog](http://about.sourcegraph.com/changelog))", IsDismissibleWithKeyValue: "months-out-of-date-7"},
			wantOnline:       &Alert{GroupValue: AlertGroupUpdate, TypeValue: AlertTypeWarning, MessageValue: "Sourcegraph is 7+ months out of date, you may be missing important security or bug fixes. Ask your site administrator to upgrade. ([changelog](http://about.sourcegraph.com/changelog))", IsDismissibleWithKeyValue: "months-out-of-date-7"},
			wantOfflineAdmin: &Alert{GroupValue: AlertGroupUpdate, TypeValue: AlertTypeError, MessageValue: "Sourcegraph is 7+ months out of date, you may be missing important security or bug fixes. A notice is shown to users. ([changelog](http://about.sourcegraph.com/changelog))"},
			wantOnlineAdmin:  &Alert{GroupValue: AlertGroupUpdate, TypeValue: AlertTypeError, MessageValue: "Sourcegraph is 7+ months out of date, you may be missing important security or bug fixes. A notice is shown to users. ([changelog](http://about.sourcegraph.com/changelog))"},
		},
		{
			name:             "13_months",
			monthsOutOfDate:  13,
			wantOffline:      &Alert{GroupValue: AlertGroupUpdate, TypeValue: AlertTypeError, MessageValue: "Sourcegraph is 13+ months out of date, you may be missing important security or bug fixes. Ask your site administrator to upgrade. ([changelog](http://about.sourcegraph.com/changelog))", IsDismissibleWithKeyValue: "months-out-of-date-13"},
			wantOnline:       &Alert{GroupValue: AlertGroupUpdate, TypeValue: AlertTypeError, MessageValue: "Sourcegraph is 13+ months out of date, you may be missing important security or bug fixes. Ask your site administrator to upgrade. ([changelog](http://about.sourcegraph.com/changelog))", IsDismissibleWithKeyValue: "months-out-of-date-13"},
			wantOfflineAdmin: &Alert{GroupValue: AlertGroupUpdate, TypeValue: AlertTypeError, MessageValue: "Sourcegraph is 13+ months out of date, you may be missing important security or bug fixes. A notice is shown to users. ([changelog](http://about.sourcegraph.com/changelog))"},
			wantOnlineAdmin:  &Alert{GroupValue: AlertGroupUpdate, TypeValue: AlertTypeError, MessageValue: "Sourcegraph is 13+ months out of date, you may be missing important security or bug fixes. A notice is shown to users. ([changelog](http://about.sourcegraph.com/changelog))"},
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
	tests := []struct {
		name string
		args AlertFuncArgs
		want []*Alert
	}{
		{
			name: "do not show anything for non-admin",
			args: AlertFuncArgs{
				IsSiteAdmin: false,
				ViewerFinalSettings: &schema.Settings{
					AlertsHideObservabilitySiteAlerts: &f,
				},
			},
			want: nil,
		},
		{
			name: "prometheus unreachable for admin",
			args: AlertFuncArgs{
				IsSiteAdmin: true,
				ViewerFinalSettings: &schema.Settings{
					AlertsHideObservabilitySiteAlerts: &f,
				},
			},
			want: []*Alert{{
				GroupValue:   AlertGroupObservability,
				TypeValue:    AlertTypeWarning,
				MessageValue: "Failed to fetch alerts status",
			}},
		},
		{
			// blocked by https://github.com/sourcegraph/sourcegraph/issues/12190
			// see observabilityActiveAlertsAlert docstrings
			name: "alerts disabled by default for admin",
			args: AlertFuncArgs{
				IsSiteAdmin: true,
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
			gotAlerts := fn(tt.args)
			if len(gotAlerts) != len(tt.want) {
				t.Errorf("expected %+v, got %+v", tt.want, gotAlerts)
				return
			}
			// test for message substring equality
			for i, got := range gotAlerts {
				want := tt.want[i]
				if got.GroupValue != want.GroupValue || got.TypeValue != want.TypeValue || got.IsDismissibleWithKeyValue != want.IsDismissibleWithKeyValue || !strings.Contains(got.MessageValue, want.MessageValue) {
					t.Errorf("expected %+v, got %+v", want, got)
				}
			}
		})
	}
}

func TestFreePlanAlert(t *testing.T) {
	plan := func(p licensing.Plan) string {
		return "plan:" + string(p)
	}

	ctx := actor.WithActor(context.Background(), actor.FromMockUser(1))
	ctx = featureflag.WithFlags(ctx, featureflag.NewMemoryStore(map[string]bool{"setup-checklist": true}, nil, nil))
	tests := []struct {
		name    string
		args    AlertFuncArgs
		license *license.Info
		want    []*Alert
	}{
		{
			name:    "do not show anything for non-admin",
			license: &license.Info{Tags: []string{plan(licensing.PlanFree0)}},
			args: AlertFuncArgs{
				IsSiteAdmin: false,
				Ctx:         ctx,
			},
			want: nil,
		},
		{
			name:    "do not show alert if license is not on free plan",
			license: &license.Info{Tags: []string{plan(licensing.PlanEnterprise0)}},
			args: AlertFuncArgs{
				IsSiteAdmin: true,
				Ctx:         ctx,
			},
			want: nil,
		},
		{
			name:    "show alert if license is on free plan 0",
			license: &license.Info{Tags: []string{plan(licensing.PlanFree0)}},
			args: AlertFuncArgs{
				IsSiteAdmin: true,
				Ctx:         ctx,
			},
			want: []*Alert{{
				GroupValue:                AlertGroupLicense,
				TypeValue:                 AlertTypeWarning,
				MessageValue:              "You're on a free Sourcegraph plan. [Upgrade](https://about.sourcegraph.com/pricing) to unlock more features and support. [Set license key](/site-admin/configuration)",
				IsDismissibleWithKeyValue: "free-plan-upgrade",
			}},
		},
		{
			name:    "show alert if license is on free plan 1",
			license: &license.Info{Tags: []string{plan(licensing.PlanFree1)}},
			args: AlertFuncArgs{
				IsSiteAdmin: true,
				Ctx:         ctx,
			},
			want: []*Alert{{
				GroupValue:                AlertGroupLicense,
				TypeValue:                 AlertTypeWarning,
				MessageValue:              "You're on a free Sourcegraph plan. [Upgrade](https://about.sourcegraph.com/pricing) to unlock more features and support. [Set license key](/site-admin/configuration)",
				IsDismissibleWithKeyValue: "free-plan-upgrade",
			}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
				return test.license, "test-signature", nil
			}
			defer func() { licensing.MockGetConfiguredProductLicenseInfo = nil }()
			gotAlerts := freePlanAlert(test.args)
			if len(gotAlerts) != len(test.want) {
				t.Errorf("expected %+v, got %+v", test.want, gotAlerts)
				return
			}
			for i, got := range gotAlerts {
				want := test.want[i]
				if diff := cmp.Diff(*want, *got); diff != "" {
					t.Fatalf("diff mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestUserCountExceededAlert(t *testing.T) {
	ctx := actor.WithActor(context.Background(), actor.FromMockUser(1))
	ctx = featureflag.WithFlags(ctx, featureflag.NewMemoryStore(map[string]bool{"setup-checklist": true}, nil, nil))

	users := dbmocks.NewMockUserStore()
	users.CountFunc.SetDefaultReturn(10, nil)
	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)

	tests := []struct {
		name    string
		args    AlertFuncArgs
		license *license.Info
		want    []*Alert
	}{
		{
			name:    "do not show anything for non-admin",
			license: &license.Info{Tags: []string{}},
			args: AlertFuncArgs{
				IsSiteAdmin: false,
				Ctx:         ctx,
				DB:          db,
			},
			want: nil,
		},
		{
			name:    "do not show alert if true up license",
			license: &license.Info{Tags: []string{licensing.TrueUpUserCountTag}, UserCount: 1},
			args: AlertFuncArgs{
				IsSiteAdmin: true,
				Ctx:         ctx,
				DB:          db,
			},
			want: nil,
		},
		{
			name:    "show alert if exceeded user count",
			license: &license.Info{Tags: []string{}, UserCount: 1},
			args: AlertFuncArgs{
				IsSiteAdmin: true,
				Ctx:         ctx,
				DB:          db,
			},
			want: []*Alert{{
				GroupValue:                AlertGroupLicense,
				TypeValue:                 AlertTypeWarning,
				MessageValue:              fmt.Sprintf("You have reached the maximum user count (%d) for your current Sourcegraph license. [Upgrade](https://about.sourcegraph.com/pricing) to support more users.", 1),
				IsDismissibleWithKeyValue: "user-count-limit",
			}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
				return test.license, "test-signature", nil
			}
			defer func() { licensing.MockGetConfiguredProductLicenseInfo = nil }()
			gotAlerts := userCountExceededAlert(test.args)
			if len(gotAlerts) != len(test.want) {
				t.Errorf("expected %+v, got %+v", test.want, gotAlerts)
				return
			}
			for i, got := range gotAlerts {
				want := test.want[i]
				if diff := cmp.Diff(*want, *got); diff != "" {
					t.Fatalf("diff mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestSMTPConfigAlert(t *testing.T) {
	type args struct {
		args AlertFuncArgs
	}

	var smtpClientError error = nil

	mockCreateSMTPClient := func(config schema.SiteConfiguration) (*smtp.Client, error) {
		return &smtp.Client{}, smtpClientError
	}

	ctx := actor.WithActor(context.Background(), actor.FromMockUser(1))
	ctx = featureflag.WithFlags(ctx, featureflag.NewMemoryStore(map[string]bool{"setup-checklist": true}, nil, nil))

	tests := []struct {
		name            string
		args            args
		config          *schema.SiteConfiguration
		smtpClientError error
		want            []*Alert
	}{
		{
			name: "do not show anything for non-admin",
			args: args{
				args: AlertFuncArgs{
					IsSiteAdmin: false,
					Ctx:         ctx,
				},
			},
			config:          nil,
			smtpClientError: nil,
			want:            nil,
		},
		{
			name: "do not show alert if smtp is configured correctly",
			args: args{
				args: AlertFuncArgs{
					IsSiteAdmin: true,
					Ctx:         ctx,
				},
			},
			config: &schema.SiteConfiguration{
				EmailSmtp: &schema.SMTPServerConfig{
					Host:           "smtp.example.com",
					Port:           567,
					Authentication: "PLAIN",
					Username:       "username",
					Password:       "password",
				},
				EmailAddress: "sourcegraph-unit-test@sourcegraph.acme.com",
			},
			smtpClientError: nil,
			want:            nil,
		},
		{
			name: "show alert if smtp config is missing",
			args: args{
				args: AlertFuncArgs{
					IsSiteAdmin: true,
					Ctx:         ctx,
				},
			},
			config: &schema.SiteConfiguration{
				EmailAddress: "sourcegraph-unit-test@sourcegraph.acme.com",
			},
			smtpClientError: nil,
			want: []*Alert{{
				GroupValue:                AlertGroupSMTP,
				TypeValue:                 AlertTypeWarning,
				MessageValue:              "SMTP is not configured. Email notifications will not be sent. [Configure SMTP](/site-admin/configuration#smtp) or [see the docs](/help/admin/config/email).",
				IsDismissibleWithKeyValue: "smtp-config-missing",
			}},
		},
		{
			name: "show alert if smtp auth config is missing",
			args: args{
				args: AlertFuncArgs{
					IsSiteAdmin: true,
					Ctx:         ctx,
				},
			},
			config: &schema.SiteConfiguration{
				EmailSmtp: &schema.SMTPServerConfig{
					Host:           "smtp.example.com",
					Port:           567,
					Authentication: "PLAIN",
				},
				EmailAddress: "sourcegraph-unit-test@sourcegraph.acme.com",
			},
			smtpClientError: nil,
			want: []*Alert{{
				GroupValue:                AlertGroupSMTP,
				TypeValue:                 AlertTypeError,
				MessageValue:              "SMTP authentication is misconfigured. SMTP Authentication is set to PLAIN, but username or password is missing. [Configure SMTP](/site-admin/configuration#smtp) or [see the docs](/help/admin/config/email).",
				IsDismissibleWithKeyValue: "smtp-config-auth-error",
			}},
		},
		{
			name: "shows 2 alerts if email address is not configured and smtp auth config is missing",
			args: args{
				args: AlertFuncArgs{
					IsSiteAdmin: true,
					Ctx:         ctx,
				},
			},
			config: &schema.SiteConfiguration{
				EmailSmtp: &schema.SMTPServerConfig{
					Host:           "smtp.example.com",
					Port:           567,
					Authentication: "PLAIN",
				},
			},
			smtpClientError: nil,
			want: []*Alert{
				{
					GroupValue:                AlertGroupSMTP,
					TypeValue:                 AlertTypeWarning,
					MessageValue:              "SMTP is not configured. Email notifications will not be sent. [Configure SMTP](/site-admin/configuration#smtp) or [see the docs](/help/admin/config/email).",
					IsDismissibleWithKeyValue: "smtp-config-missing",
				}, {
					GroupValue:                AlertGroupSMTP,
					TypeValue:                 AlertTypeError,
					MessageValue:              "SMTP authentication is misconfigured. SMTP Authentication is set to PLAIN, but username or password is missing. [Configure SMTP](/site-admin/configuration#smtp) or [see the docs](/help/admin/config/email).",
					IsDismissibleWithKeyValue: "smtp-config-auth-error",
				}},
		},
		{
			name: "show alert if smtp client cannot reach smtp server",
			args: args{
				args: AlertFuncArgs{
					IsSiteAdmin: true,
					Ctx:         ctx,
				},
			},
			config: &schema.SiteConfiguration{
				EmailSmtp: &schema.SMTPServerConfig{
					Host:           "smtp.example.com",
					Port:           567,
					Authentication: "PLAIN",
					Username:       "username",
					Password:       "password",
				},
				EmailAddress: "sourcegraph-unit-test@sourcegraph.acme.com",
			},
			smtpClientError: errors.Newf("test smtp client cannot reach smtp server"),
			want: []*Alert{{
				GroupValue:                AlertGroupSMTP,
				TypeValue:                 AlertTypeError,
				MessageValue:              "SMTP server cannot be reached, please check your SMTP configuration. [Configure SMTP](/site-admin/configuration#smtp) or [see the docs](/help/admin/config/email).",
				IsDismissibleWithKeyValue: "smtp-client-error",
			}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			smtpClientError = test.smtpClientError
			if test.config != nil {
				conf.Mock(&conf.Unified{SiteConfiguration: *test.config})
			}

			defer func() {
				smtpClientError = nil
				conf.Mock(nil)
			}()

			gotAlerts := smtpConfigAlert(mockCreateSMTPClient)(test.args.args)
			if len(gotAlerts) != len(test.want) {
				t.Errorf("got %s", gotAlerts[0].Message())
				t.Errorf("expected %+v, got %+v", test.want, gotAlerts)
				return
			}
			for i, got := range gotAlerts {
				want := test.want[i]
				if diff := cmp.Diff(*want, *got); diff != "" {
					t.Fatalf("diff mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestEmptyExternalURLAlert(t *testing.T) {
	type args struct {
		args AlertFuncArgs
	}

	ctx := actor.WithActor(context.Background(), actor.FromMockUser(1))
	ctx = featureflag.WithFlags(ctx, featureflag.NewMemoryStore(map[string]bool{"setup-checklist": true}, nil, nil))

	tests := []struct {
		name   string
		args   args
		config *schema.SiteConfiguration
		want   []*Alert
	}{
		{
			name: "do not show anything for non-admin",
			args: args{
				args: AlertFuncArgs{
					IsSiteAdmin: false,
					Ctx:         ctx,
				},
			},
			config: nil,
			want:   nil,
		},
		{
			name: "do not show alert if externalURL is configured correctly",
			args: args{
				args: AlertFuncArgs{
					IsSiteAdmin: true,
					Ctx:         ctx,
				},
			},
			config: &schema.SiteConfiguration{
				ExternalURL: "https://sourcegraph.example.test",
			},
			want: nil,
		},
		{
			name: "show alert if externalURL is empty",
			args: args{
				args: AlertFuncArgs{
					IsSiteAdmin: true,
					Ctx:         ctx,
				},
			},
			config: &schema.SiteConfiguration{
				ExternalURL: "",
			},
			want: []*Alert{{
				GroupValue:   AlertGroupExternalURL,
				TypeValue:    AlertTypeError,
				MessageValue: "`externalURL` is required to be set for many features of Sourcegraph to work correctly.",
			}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.config != nil {
				conf.Mock(&conf.Unified{SiteConfiguration: *test.config})
			}

			defer func() {
				conf.Mock(nil)
			}()

			gotAlerts := emptyExternalURLAlert(test.args.args)
			if len(gotAlerts) != len(test.want) {
				t.Errorf("got %s", gotAlerts[0].Message())
				t.Errorf("expected %+v, got %+v", test.want, gotAlerts)
				return
			}
			for i, got := range gotAlerts {
				want := test.want[i]
				if diff := cmp.Diff(*want, *got); diff != "" {
					t.Fatalf("diff mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
