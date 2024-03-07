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
	"github.com/sourcegraph/sourcegraph/internal/ssc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type mockSSCValue struct {
	subscription  *ssc.Subscription
	samsAccountID string
}

type mockSSCClient struct {
	t              *testing.T
	mockSSCValue   []mockSSCValue
	shouldBeCalled bool
}

func isExpectedID(id string, expected []string) bool {
	for _, expectedID := range expected {
		if id == expectedID {
			return true
		}
	}

	return false
}

func (m *mockSSCClient) FetchSubscriptionBySAMSAccountID(
	ctx context.Context, samsAccountID string) (*ssc.Subscription, error) {
	if !m.shouldBeCalled {
		m.t.Error("FetchSubscriptionBySAMSAccountID should not have be called")
	}
	assert.NotNil(m.t, ctx)

	for _, v := range m.mockSSCValue {
		if v.samsAccountID == samsAccountID {
			return v.subscription, nil
		}
	}

	m.t.Errorf("FetchSubscriptionBySAMSAccountID should not have be called with the given samsAccountID: %s", samsAccountID)
	return nil, nil
}

func toTimePtr(t time.Time) *time.Time {
	return &t
}

func TestGetSubscriptionForUser(t *testing.T) {
	samsAccountIDWithSubscription := "having-subscription"
	samsAccountIDWithoutSubscription := "no-subscription"

	tests := []struct {
		name                 string
		user                 types.User
		today                time.Time
		mockSSC              []mockSSCValue
		expectedSubscription UserSubscription
	}{
		{
			name: "user without SAMS account & SSC subscription",
			user: types.User{
				ID:        1,
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			today:   time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
			mockSSC: []mockSSCValue{},
			expectedSubscription: UserSubscription{
				Status:               ssc.SubscriptionStatusActive,
				Plan:                 UserSubscriptionPlanFree,
				ApplyProRateLimits:   false,
				CurrentPeriodStartAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				CurrentPeriodEndAt:   time.Date(2024, 1, 31, 23, 59, 59, 59, time.UTC),
			},
		},
		{
			name: "user with SAMS account without SSC subscription",
			user: types.User{
				ID:        1,
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			today:   time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
			mockSSC: []mockSSCValue{{samsAccountID: samsAccountIDWithoutSubscription}},
			expectedSubscription: UserSubscription{
				Status:               ssc.SubscriptionStatusActive,
				Plan:                 UserSubscriptionPlanFree,
				ApplyProRateLimits:   false,
				CurrentPeriodStartAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				CurrentPeriodEndAt:   time.Date(2024, 1, 31, 23, 59, 59, 59, time.UTC),
			},
		},
		{
			name: "user with SAMS account & SSC subscription",
			user: types.User{
				ID:        1,
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			today: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
			mockSSC: []mockSSCValue{{
				subscription: &ssc.Subscription{
					Status:             ssc.SubscriptionStatusActive,
					BillingInterval:    ssc.BillingIntervalMonthly,
					CancelAtPeriodEnd:  false,
					CurrentPeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339Nano),
					CurrentPeriodEnd:   time.Date(2024, 1, 31, 23, 59, 59, 59, time.UTC).Format(time.RFC3339Nano),
				},
				samsAccountID: samsAccountIDWithSubscription,
			}},
			expectedSubscription: UserSubscription{
				Status:               ssc.SubscriptionStatusActive,
				Plan:                 UserSubscriptionPlanPro,
				ApplyProRateLimits:   true,
				CurrentPeriodStartAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				CurrentPeriodEndAt:   time.Date(2024, 1, 31, 23, 59, 59, 59, time.UTC),
			},
		},
		{
			// to cover the bug where 1 dotcom user can have mutliple SAMS identities linked
			name: "user with multiple SAMS account but 1 SSC subscription",
			user: types.User{
				ID:        1,
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			today: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
			mockSSC: []mockSSCValue{{
				samsAccountID: samsAccountIDWithoutSubscription,
			}, {
				subscription: &ssc.Subscription{
					Status:             ssc.SubscriptionStatusActive,
					BillingInterval:    ssc.BillingIntervalMonthly,
					CancelAtPeriodEnd:  false,
					CurrentPeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339Nano),
					CurrentPeriodEnd:   time.Date(2024, 1, 31, 23, 59, 59, 59, time.UTC).Format(time.RFC3339Nano),
				},
				samsAccountID: samsAccountIDWithSubscription,
			}},
			expectedSubscription: UserSubscription{
				Status:               ssc.SubscriptionStatusActive,
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

			// Mock out the lookup for the user's external identities, where we fetch their SAMS account ID.
			db := dbmocks.NewMockDB()
			userExternalAccount := dbmocks.NewMockUserExternalAccountsStore()
			userExternalAccount.ListFunc.SetDefaultHook(func(ctx context.Context, opts database.ExternalAccountsListOptions) ([]*extsvc.Account, error) {
				assert.Equal(t, test.user.ID, opts.UserID)
				var accounts []*extsvc.Account

				for _, mockSSCValue := range test.mockSSC {
					accounts = append(accounts, &extsvc.Account{AccountSpec: extsvc.AccountSpec{
						AccountID:   mockSSCValue.samsAccountID,
						ServiceType: "openidconnect",
						ServiceID:   fmt.Sprintf("https://%s/", ssc.GetSAMSHostName()),
					}})
				}

				return accounts, nil
			})
			db.UserExternalAccountsFunc.SetDefaultReturn(userExternalAccount)

			expectToBeCalled := len(test.mockSSC) != 0

			getSSCClient = func() (ssc.Client, error) {
				return &mockSSCClient{
					t:              t,
					mockSSCValue:   test.mockSSC,
					shouldBeCalled: expectToBeCalled,
				}, nil
			}

			actualSubscription, err := SubscriptionForUser(ctx, db, test.user)
			assert.NoError(t, err)
			assert.NotNil(t, actualSubscription)
			assert.Equal(t, test.expectedSubscription, *actualSubscription)
		})
	}
}
