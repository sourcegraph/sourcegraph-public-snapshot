package licenseexpiration

import (
	"context"
	"testing"
	"time"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	subscriptions "github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/subscriptions"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/utctime"
	"github.com/sourcegraph/sourcegraph/internal/slack"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestHandle(t *testing.T) {
	mockTime := time.Date(2024, 7, 8, 16, 39, 16, 4277000, time.Local)

	store := NewMockStore()
	store.NowFunc.SetDefaultReturn(mockTime)
	store.EnvFunc.SetDefaultReturn("dev")
	store.TryAcquireJobFunc.SetDefaultReturn(true, nil, nil)
	store.ListSubscriptionsFunc.SetDefaultReturn(
		[]*subscriptions.SubscriptionWithConditions{
			{Subscription: subscriptions.Subscription{
				ID:                       "e9450fb2-87c7-47ae-a713-a376c4618faa",
				SalesforceSubscriptionID: pointers.Ptr("sf-sub-id"),
			}},
			{Subscription: subscriptions.Subscription{
				ID:          "26136564-b319-4be4-98ff-7b8710abf4af",
				DisplayName: pointers.Ptr("My Special Subscription"),
			}},
			{Subscription: subscriptions.Subscription{
				ID: "32bda851-5761-4b18-81bf-d20f39bd5cb6",
			}},
		},
		nil)
	store.GetActiveLicenseFunc.SetDefaultHook(func(ctx context.Context, sub string) (*subscriptions.LicenseWithConditions, error) {
		switch sub {
		case "e9450fb2-87c7-47ae-a713-a376c4618faa":
			return &subscriptions.LicenseWithConditions{
				SubscriptionLicense: subscriptions.SubscriptionLicense{
					ExpireAt: utctime.FromTime(mockTime.Add((24 + 1) * time.Hour)), // day away
				},
			}, nil
		case "26136564-b319-4be4-98ff-7b8710abf4af":
			return &subscriptions.LicenseWithConditions{
				SubscriptionLicense: subscriptions.SubscriptionLicense{
					ExpireAt: utctime.FromTime(mockTime.Add((7*24 + 1) * time.Hour)), // week away
				},
			}, nil
		default:
			return &subscriptions.LicenseWithConditions{
				SubscriptionLicense: subscriptions.SubscriptionLicense{
					ExpireAt: utctime.FromTime(mockTime.Add((99 * 24) * time.Hour)), // far away time
				},
			}, nil
		}
	})

	h := handler{
		logger:                  logtest.Scoped(t),
		store:                   store,
		licenseCheckConcurrency: 1,
	}

	err := h.Handle(context.Background())
	require.NoError(t, err)

	var payloads []*slack.Payload
	for _, call := range store.PostToSlackFunc.History() {
		payloads = append(payloads, call.Arg1)
	}
	autogold.Expect([]*slack.Payload{
		{
			Text: "The active license for subscription *es_e9450fb2-87c7-47ae-a713-a376c4618faa* (Salesforce subscription: `sf-sub-id`) <https://sourcegraph.com/site-admin/dotcom/product/subscriptions/es_e9450fb2-87c7-47ae-a713-a376c4618faa?env=dev|will expire *in the next 24 hours*> :rotating_light:",
		},
		{Text: "The active license for subscription *My Special Subscription* (Salesforce subscription: `unknown`) <https://sourcegraph.com/site-admin/dotcom/product/subscriptions/es_26136564-b319-4be4-98ff-7b8710abf4af?env=dev|will expire *in 7 days*>"},
	}).Equal(t, payloads)
}
