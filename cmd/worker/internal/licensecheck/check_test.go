package licensecheck

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/derision-test/glock"
	mockrequire "github.com/derision-test/go-mockgen/v2/testutil/require"
	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"

	subscriptionlicensechecksv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptionlicensechecks/v1"
)

func Test_calcDurationToWaitForNextHandle(t *testing.T) {
	kv := rcache.SetupForTest(t)

	cleanupStore := func() {
		_ = kv.Del(licensing.LicenseValidityStoreKey)
		_ = kv.Del(lastCalledAtStoreKey)
	}

	now := time.Now().Round(time.Second)
	clock := glock.NewMockClock()
	clock.SetCurrent(now)

	tests := map[string]struct {
		lastCalledAt string
		want         time.Duration
		wantErr      bool
	}{
		"returns 0 if last called at is empty": {
			lastCalledAt: "",
			want:         0,
			wantErr:      true,
		},
		"returns 0 if last called at is invalid": {
			lastCalledAt: "invalid",
			want:         0,
			wantErr:      true,
		},
		"returns 0 if last called at is in the future": {
			lastCalledAt: now.Add(time.Minute).Format(time.RFC3339),
			want:         0,
			wantErr:      true,
		},
		"returns 0 if last called at is before licensing.LicenseCheckInterval": {
			lastCalledAt: now.Add(-licensing.LicenseCheckInterval - time.Minute).Format(time.RFC3339),
			want:         0,
			wantErr:      false,
		},
		"returns 0 if last called at is at licensing.LicenseCheckInterval": {
			lastCalledAt: now.Add(-licensing.LicenseCheckInterval).Format(time.RFC3339),
			want:         0,
			wantErr:      false,
		},
		"returns diff between last called at and now": {
			lastCalledAt: now.Add(-time.Hour).Format(time.RFC3339),
			want:         licensing.LicenseCheckInterval - time.Hour,
			wantErr:      false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cleanupStore()
			if test.lastCalledAt != "" {
				_ = kv.Set(lastCalledAtStoreKey, test.lastCalledAt)
			}

			got, err := calcDurationSinceLastCalled(kv, clock)
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.want, got)
		})
	}
}

func Test_licenseChecker(t *testing.T) {
	kv := rcache.SetupForTest(t)

	cleanupStore := func() {
		_ = kv.Del(licensing.LicenseValidityStoreKey)
		_ = kv.Del(lastCalledAtStoreKey)
	}

	siteID := "some-site-id"

	skipTests := map[string]struct {
		license *license.Info
	}{
		"skips check if license is air gapped": {
			license: &license.Info{
				Tags: []string{string(licensing.FeatureAllowAirGapped)},
			},
		},
		"skips check on free license": {
			license: &licensing.GetFreeLicenseInfo().Info,
		},
	}

	for name, test := range skipTests {
		t.Run(name, func(t *testing.T) {
			cleanupStore()
			defaultMockGetLicense := licensing.MockGetConfiguredProductLicenseInfo
			licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
				return test.license, "", nil
			}

			t.Cleanup(func() {
				licensing.MockGetConfiguredProductLicenseInfo = defaultMockGetLicense
			})

			mockDB := dbmocks.NewMockDB()
			gs := dbmocks.NewMockGlobalStateStore()
			mockDB.GlobalStateFunc.SetDefaultReturn(gs)
			checks := NewMockSubscriptionLicenseChecksServiceClient()
			gs.GetFunc.SetDefaultReturn(database.GlobalState{
				SiteID: siteID,
			}, nil)
			handler := licenseChecker{
				db:     mockDB,
				checks: checks,
				logger: logtest.NoOp(t),
				kv:     kv,
			}

			err := handler.Handle(context.Background())
			require.NoError(t, err)

			// check doer NOT called
			mockrequire.NotCalled(t, checks.CheckLicenseKeyFunc)

			// check result was set to true
			valid, err := kv.Get(licensing.LicenseValidityStoreKey).Bool()
			require.NoError(t, err)
			require.True(t, valid)

			// check last called at was set
			lastCalledAt, err := kv.Get(lastCalledAtStoreKey).String()
			require.NoError(t, err)
			require.NotEmpty(t, lastCalledAt)
		})
	}

	tests := map[string]struct {
		response      *subscriptionlicensechecksv1.CheckLicenseKeyResponse
		responseError error

		wantValid bool
		wantError autogold.Value
		baseUrl   *string
		reason    *string
	}{
		"returns error if unexpected error": {
			responseError: errors.New("unexpected error"),
			wantError:     autogold.Expect("unexpected error"),
		},
		"is invalid if error CodeNotFound": {
			responseError: connect.NewError(connect.CodeNotFound, errors.New("not found")),
			wantValid:     false,
		},
		`returns correct result for "true"`: {
			response: &subscriptionlicensechecksv1.CheckLicenseKeyResponse{
				Valid: true,
			},
			wantValid: true,
		},
		`returns correct result for "false"`: {
			response: &subscriptionlicensechecksv1.CheckLicenseKeyResponse{
				Valid:  false,
				Reason: "some reason",
			},
			wantValid: false,
			reason:    pointers.Ptr("some reason"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cleanupStore()
			defaultMockGetLicense := licensing.MockGetConfiguredProductLicenseInfo
			licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
				return &license.Info{Tags: []string{"plan:enterprise-0"}}, "", nil
			}
			t.Cleanup(func() {
				licensing.MockGetConfiguredProductLicenseInfo = defaultMockGetLicense
			})

			confClient := conf.MockClient()
			confClient.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{
				LicenseKey: "license-key",
			}})

			checks := NewMockSubscriptionLicenseChecksServiceClient()
			checks.CheckLicenseKeyFunc.SetDefaultReturn(
				connect.NewResponse(test.response),
				test.responseError,
			)
			mockDB := dbmocks.NewMockDB()
			gs := dbmocks.NewMockGlobalStateStore()
			mockDB.GlobalStateFunc.SetDefaultReturn(gs)
			gs.GetFunc.SetDefaultReturn(database.GlobalState{
				SiteID: siteID,
			}, nil)
			checker := licenseChecker{
				db:     mockDB,
				checks: checks,
				logger: logtest.NoOp(t),
				kv:     kv,
				conf:   confClient,
			}

			err := checker.Handle(context.Background())
			if test.wantError != nil {
				require.Error(t, err)
				test.wantError.Equal(t, err.Error())

				// check result was NOT set
				require.True(t, kv.Get(licensing.LicenseValidityStoreKey).IsNil())
			} else {
				require.NoError(t, err)

				// check result was set
				got, err := kv.Get(licensing.LicenseValidityStoreKey).Bool()
				require.NoError(t, err)
				require.Equal(t, test.wantValid, got)

				// check result reason was set
				if test.reason != nil {
					got, err := kv.Get(licensing.LicenseInvalidReason).String()
					require.NoError(t, err)
					require.Equal(t, *test.reason, got)
				}
			}

			// check last called at was set if client did not error
			if test.responseError == nil {
				lastCalledAt, err := kv.Get(lastCalledAtStoreKey).String()
				require.NoError(t, err)
				require.NotEmpty(t, lastCalledAt)
			}

			// check doer with proper parameters
			mockrequire.Called(t, checks.CheckLicenseKeyFunc)

			// The token for the license.
			args := checks.CheckLicenseKeyFunc.History()[0].Arg1
			require.Equal(t, "license-key", args.Msg.LicenseKey)
			require.Equal(t, siteID, args.Msg.InstanceId)
		})
	}
}
