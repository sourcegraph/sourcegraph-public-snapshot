package cody

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/ssc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type mockSSCClient struct {
	t                     *testing.T
	expectedSAMSAccountID *string
	mockSSCSubscription   *ssc.Subscription
	shouldBeCalled        bool
	Called                bool
}

func (m *mockSSCClient) FetchSubscriptionBySAMSAccountID(
	ctx context.Context, samsAccountID string) (*ssc.Subscription, error) {
	if !m.shouldBeCalled {
		m.t.Error("FetchSubscriptionBySAMSAccountID should not have be called")
	}
	assert.NotNil(m.t, ctx)
	assert.NotNil(m.t, m.expectedSAMSAccountID)
	assert.Equal(m.t, *m.expectedSAMSAccountID, samsAccountID)

	m.Called = true
	return m.mockSSCSubscription, nil
}

func toTimePtr(t time.Time) *time.Time {
	return &t
}

func TestGetSubscriptionForUser(t *testing.T) {
	testSAMSAccountID := "123"

	tests := []struct {
		name                         string
		user                         types.User
		today                        time.Time
		mockSSCSubscription          *ssc.Subscription
		mockSAMSAccountID            *string
		useSSCFeatureFlag            bool
		codyProTrialEndedFeatureFlag bool
		expectedSubscription         UserSubscription
	}{
		{
			name: "free user without SAMS account & SSC subscription",
			user: types.User{
				ID:               1,
				CreatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				CodyProEnabledAt: nil,
			},
			today:               time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
			mockSSCSubscription: nil,
			mockSAMSAccountID:   nil,
			useSSCFeatureFlag:   true,
			expectedSubscription: UserSubscription{
				Status:               ssc.SubscriptionStatusPending,
				Plan:                 UserSubscriptionPlanFree,
				ApplyProRateLimits:   false,
				CurrentPeriodStartAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				CurrentPeriodEndAt:   time.Date(2024, 1, 31, 23, 59, 59, 59, time.UTC),
			},
		},
		{
			name: "free user with SAMS account without SSC subscription",
			user: types.User{
				ID:               1,
				CreatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				CodyProEnabledAt: nil,
			},
			today:               time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
			mockSSCSubscription: nil,
			mockSAMSAccountID:   &testSAMSAccountID,
			useSSCFeatureFlag:   true,
			expectedSubscription: UserSubscription{
				Status:               ssc.SubscriptionStatusPending,
				Plan:                 UserSubscriptionPlanFree,
				ApplyProRateLimits:   false,
				CurrentPeriodStartAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				CurrentPeriodEndAt:   time.Date(2024, 1, 31, 23, 59, 59, 59, time.UTC),
			},
		},
		{
			// Possible when the user opted for Pro only after the Feb release.
			name: "free user with SAMS account & SSC subscription",
			user: types.User{
				ID:               1,
				CreatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				CodyProEnabledAt: nil,
			},
			today: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
			mockSSCSubscription: &ssc.Subscription{
				Status:             ssc.SubscriptionStatusActive,
				BillingInterval:    ssc.BillingIntervalMonthly,
				CancelAtPeriodEnd:  false,
				CurrentPeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339Nano),
				CurrentPeriodEnd:   time.Date(2024, 1, 31, 23, 59, 59, 59, time.UTC).Format(time.RFC3339Nano),
			},
			mockSAMSAccountID: &testSAMSAccountID,
			useSSCFeatureFlag: true,
			expectedSubscription: UserSubscription{
				Status:               ssc.SubscriptionStatusActive,
				Plan:                 UserSubscriptionPlanPro,
				ApplyProRateLimits:   true,
				CurrentPeriodStartAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				CurrentPeriodEndAt:   time.Date(2024, 1, 31, 23, 59, 59, 59, time.UTC),
			},
		},
		{
			// possible when the user opted for Pro only after the feb release
			name: "free user with SAMS account & SSC subscription but feature flag disabled",
			user: types.User{
				ID:               1,
				CreatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				CodyProEnabledAt: nil,
			},
			today: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
			mockSSCSubscription: &ssc.Subscription{
				Status:             ssc.SubscriptionStatusActive,
				BillingInterval:    ssc.BillingIntervalMonthly,
				CancelAtPeriodEnd:  false,
				CurrentPeriodStart: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339Nano),
				CurrentPeriodEnd:   time.Date(2025, 1, 31, 23, 59, 59, 59, time.UTC).Format(time.RFC3339Nano),
			},
			mockSAMSAccountID: &testSAMSAccountID,
			useSSCFeatureFlag: false,
			expectedSubscription: UserSubscription{
				Status:               ssc.SubscriptionStatusPending,
				Plan:                 UserSubscriptionPlanFree,
				ApplyProRateLimits:   false,
				CurrentPeriodStartAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				CurrentPeriodEndAt:   time.Date(2024, 1, 31, 23, 59, 59, 59, time.UTC),
			},
		},
		{
			name: "pro user without SAMS account & SSC subscription before release",
			user: types.User{
				ID:               1,
				CreatedAt:        time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				CodyProEnabledAt: toTimePtr(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
			today:               time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
			mockSSCSubscription: nil,
			mockSAMSAccountID:   nil,
			useSSCFeatureFlag:   true,
			expectedSubscription: UserSubscription{
				Status:               ssc.SubscriptionStatusPending,
				Plan:                 UserSubscriptionPlanPro,
				ApplyProRateLimits:   true,
				CurrentPeriodStartAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				CurrentPeriodEndAt:   time.Date(2024, 1, 31, 23, 59, 59, 59, time.UTC),
			},
		},
		{
			name: "pro user without SAMS account & SSC subscription after release",
			user: types.User{
				ID:               1,
				CreatedAt:        time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				CodyProEnabledAt: toTimePtr(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
			today:                        time.Date(2024, 2, 16, 0, 0, 0, 0, time.UTC),
			mockSSCSubscription:          nil,
			mockSAMSAccountID:            nil,
			useSSCFeatureFlag:            true,
			codyProTrialEndedFeatureFlag: true,
			expectedSubscription: UserSubscription{
				Status:               ssc.SubscriptionStatusPending,
				Plan:                 UserSubscriptionPlanPro,
				ApplyProRateLimits:   false,
				CurrentPeriodStartAt: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
				CurrentPeriodEndAt:   time.Date(2024, 2, 29, 23, 59, 59, 59, time.UTC),
			},
		},
		{
			name: "pro user with SAMS account without SSC subscription before release",
			user: types.User{
				ID:               1,
				CreatedAt:        time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				CodyProEnabledAt: toTimePtr(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
			today:               time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
			mockSSCSubscription: nil,
			mockSAMSAccountID:   &testSAMSAccountID,
			useSSCFeatureFlag:   true,
			expectedSubscription: UserSubscription{
				Status:               ssc.SubscriptionStatusPending,
				Plan:                 UserSubscriptionPlanPro,
				ApplyProRateLimits:   true,
				CurrentPeriodStartAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				CurrentPeriodEndAt:   time.Date(2024, 1, 31, 23, 59, 59, 59, time.UTC),
			},
		},
		{
			name: "pro user with SAMS account without SSC subscription after release",
			user: types.User{
				ID:               1,
				CreatedAt:        time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				CodyProEnabledAt: toTimePtr(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)),
			},
			today:                        time.Date(2024, 2, 16, 0, 0, 0, 0, time.UTC),
			mockSSCSubscription:          nil,
			mockSAMSAccountID:            &testSAMSAccountID,
			useSSCFeatureFlag:            true,
			codyProTrialEndedFeatureFlag: true,
			expectedSubscription: UserSubscription{
				Status:               ssc.SubscriptionStatusPending,
				Plan:                 UserSubscriptionPlanPro,
				ApplyProRateLimits:   false,
				CurrentPeriodStartAt: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
				CurrentPeriodEndAt:   time.Date(2024, 2, 29, 23, 59, 59, 59, time.UTC),
			},
		},
		{
			name: "pro user with SAMS account & SSC subscription",
			user: types.User{
				ID:               1,
				CreatedAt:        time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				CodyProEnabledAt: toTimePtr(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
			today: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
			mockSSCSubscription: &ssc.Subscription{
				Status:             ssc.SubscriptionStatusActive,
				BillingInterval:    ssc.BillingIntervalMonthly,
				CancelAtPeriodEnd:  false,
				CurrentPeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339Nano),
				CurrentPeriodEnd:   time.Date(2024, 1, 31, 23, 59, 59, 59, time.UTC).Format(time.RFC3339Nano),
			},
			mockSAMSAccountID: &testSAMSAccountID,
			useSSCFeatureFlag: true,
			expectedSubscription: UserSubscription{
				Status:               ssc.SubscriptionStatusActive,
				Plan:                 UserSubscriptionPlanPro,
				ApplyProRateLimits:   true,
				CurrentPeriodStartAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				CurrentPeriodEndAt:   time.Date(2024, 1, 31, 23, 59, 59, 59, time.UTC),
			},
		},
		{
			// possible when the user opted for Pro only after the feb release
			name: "pro user with SAMS account & SSC subscription but feature flag disabled before release",
			user: types.User{
				ID:               1,
				CreatedAt:        time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				CodyProEnabledAt: toTimePtr(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
			today: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
			mockSSCSubscription: &ssc.Subscription{
				Status:             ssc.SubscriptionStatusActive,
				BillingInterval:    ssc.BillingIntervalMonthly,
				CancelAtPeriodEnd:  false,
				CurrentPeriodStart: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339Nano),
				CurrentPeriodEnd:   time.Date(2025, 1, 31, 23, 59, 59, 59, time.UTC).Format(time.RFC3339Nano),
			},
			mockSAMSAccountID: &testSAMSAccountID,
			useSSCFeatureFlag: false,
			expectedSubscription: UserSubscription{
				Status:               ssc.SubscriptionStatusPending,
				Plan:                 UserSubscriptionPlanPro,
				ApplyProRateLimits:   true,
				CurrentPeriodStartAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				CurrentPeriodEndAt:   time.Date(2024, 1, 31, 23, 59, 59, 59, time.UTC),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := actor.WithActor(context.Background(), actor.FromUser(test.user.ID))
			ctx = withCurrentTimeMock(ctx, test.today)

			// Mock out feature flags.
			flags := map[string]bool{
				featureFlagUseSCCForSubscription: test.useSSCFeatureFlag,
				featureFlagCodyProTrialEnded:     test.codyProTrialEndedFeatureFlag,
			}
			ctx = featureflag.WithFlags(ctx, featureflag.NewMemoryStore(flags, flags, flags))

			// Mock out the lookup for the user's external identities, where we fetch their SAMS account ID.
			db := dbmocks.NewMockDB()
			userExternalAccount := dbmocks.NewMockUserExternalAccountsStore()
			userExternalAccount.ListFunc.SetDefaultHook(func(ctx context.Context, opts database.ExternalAccountsListOptions) ([]*extsvc.Account, error) {
				assert.Equal(t, test.user.ID, opts.UserID)
				if test.mockSAMSAccountID != nil {
					// The code identifies the SAMS account by checking the service type and service ID.
					// So we need to set these in the values returned from the mock.
					samsAccountSpec := extsvc.AccountSpec{
						AccountID:   *test.mockSAMSAccountID,
						ServiceType: "openidconnect",
						ServiceID:   fmt.Sprintf("https://%s/", ssc.SAMSProdHostname),
					}
					return []*extsvc.Account{
						{
							AccountSpec: samsAccountSpec,
						},
					}, nil
				}
				return []*extsvc.Account{}, nil
			})
			db.UserExternalAccountsFunc.SetDefaultReturn(userExternalAccount)

			expectToBeCalled := test.useSSCFeatureFlag && test.mockSAMSAccountID != nil

			getSSCClient = func() (ssc.Client, error) {
				return &mockSSCClient{
					t:                     t,
					expectedSAMSAccountID: test.mockSAMSAccountID,
					mockSSCSubscription:   test.mockSSCSubscription,
					shouldBeCalled:        expectToBeCalled,
				}, nil
			}

			actualSubscription, err := SubscriptionForUser(ctx, db, test.user)
			assert.NoError(t, err)
			assert.NotNil(t, actualSubscription)
			assert.Equal(t, test.expectedSubscription, *actualSubscription)
		})
	}
}
