package productsubscription

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/slack"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestMaybeCheckAnomalies(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbmocks.NewMockDB()

	mockClient := &fakeSlackClient{}

	mockClock := glock.NewMockClock()

	rs := redispool.NewMockKeyValue()

	testCases := []struct {
		name      string
		lastCheck time.Time
		hasCalled bool
	}{
		{
			name:      "no previous check time",
			lastCheck: time.Time{},
			hasCalled: true,
		},
		{
			name:      "previous check time within 24 hours",
			lastCheck: mockClock.Now().UTC().Add(-23 * time.Hour),
			hasCalled: false,
		},
		{
			name:      "previous check time over 24 hours",
			lastCheck: mockClock.Now().UTC().Add(-25 * time.Hour),
			hasCalled: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			rs.SetFunc.SetDefaultHook(func(string, interface{}) error {
				called = true
				return nil
			})
			rs.GetFunc.SetDefaultHook(func(string) redispool.Value {
				if tc.lastCheck.IsZero() {
					return redispool.NewValue(nil, redis.ErrNil)
				}
				return redispool.NewValue(tc.lastCheck.Format(time.RFC3339), nil)
			})

			maybeCheckAnomalies(logger, db, mockClient, mockClock, rs)

			require.Equal(t, tc.hasCalled, called)
		})
	}
}

func TestCheckAnomalies(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	clock := glock.NewMockClock()
	clock.SetCurrent(time.Unix(1686666666, 0)) // 2023-06-13T14:31:06Z

	siteID := "02a5a9e6-b45e-4e1a-b2a0-f812620e6dff"
	licenseID := "22e0cc8e-57ad-4dd9-be54-0f94d6e9964d"

	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			Dotcom: &schema.Dotcom{
				SlackLicenseAnomallyWebhook: "https://slack.com/webhook",
			},
			ExternalURL: "https://sourcegraph.acme.com",
		},
	})

	sub1ID := "e9450fb2-87c7-47ae-a713-a376c4618faa"
	sub2ID := "26136564-b319-4be4-98ff-7b8710abf4af"
	mocks.subscriptions.List = func(ctx context.Context, opt dbSubscriptionsListOptions) ([]*dbSubscription, error) {
		return []*dbSubscription{
			{ID: sub1ID},
			{ID: sub2ID},
		}, nil
	}
	mocks.licenses.List = func(ctx context.Context, opt dbLicensesListOptions) ([]*dbLicense, error) {
		if opt.ProductSubscriptionID == sub2ID {
			return []*dbLicense{{ID: licenseID, LicenseKey: "key", ProductSubscriptionID: opt.ProductSubscriptionID, SiteID: &siteID, LicenseVersion: pointers.Ptr(int32(2))}}, nil
		}
		return []*dbLicense{}, nil
	}

	t.Cleanup(func() {
		conf.Mock(nil)
		mocks.subscriptions = mockSubscriptions{}
		mocks.licenses = mockLicenses{}
		// licensing.MockParseProductLicenseKeyWithBuiltinOrGenerationKey = nil
	})

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	eventJSON, err := json.Marshal(struct {
		SiteID string `json:"site_id,omitempty"`
	}{
		SiteID: siteID,
	})
	require.NoError(t, err)

	cleanupDB := func(t *testing.T) {
		t.Helper()

		if t.Failed() {
			return
		}
		_, err := db.Handle().QueryContext(ctx, `TRUNCATE event_logs`)
		require.NoError(t, err)
	}

	createEvents := func(t *testing.T, times []time.Time) {
		t.Helper()

		if len(times) == 0 {
			return
		}

		events := make([]*database.Event, len(times))
		for i, ts := range times {
			events[i] = &database.Event{
				Name:            EventNameSuccess,
				URL:             "",
				AnonymousUserID: "backend",
				Argument:        eventJSON,
				Source:          "BACKEND",
				Timestamp:       ts,
			}
		}
		//lint:ignore SA1019 existing usage of deprecated functionality.
		// Use EventRecorder from internal/telemetryrecorder instead.
		err = db.EventLogs().BulkInsert(ctx, events)
		require.NoError(t, err)
	}

	slackMessage := fmt.Sprintf(slackMessageFmt, "https://sourcegraph.acme.com", url.QueryEscape(sub2ID), url.QueryEscape(licenseID), licenseID, siteID)

	tests := []struct {
		name      string
		times     []time.Time
		anomalous bool
	}{
		{
			name:      "no events",
			times:     []time.Time{},
			anomalous: false,
		},
		{
			name: "ok time interval between events",
			times: []time.Time{
				clock.Now().Add(-40 * time.Hour),
				clock.Now().Add(-28 * time.Hour),
				clock.Now().Add(-24 * time.Hour), // mimics redis cleanup and instance restart
				clock.Now().Add(-12 * time.Hour),
			},
			anomalous: false,
		},
		{
			name: "Two instances sending events",
			times: []time.Time{
				clock.Now().Add(-40 * time.Hour),
				clock.Now().Add(-29 * time.Hour),
				clock.Now().Add(-28 * time.Hour),
				clock.Now().Add(-17 * time.Hour),
				clock.Now().Add(-16 * time.Hour),
				clock.Now().Add(-5 * time.Hour),
				clock.Now().Add(-4 * time.Hour),
			},
			anomalous: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Cleanup(func() {
				cleanupDB(t)
			})

			createEvents(t, test.times)

			wantPayloads := []*slack.Payload(nil)
			if test.anomalous {
				wantPayloads = []*slack.Payload{{Text: slackMessage}}
			}

			client := &fakeSlackClient{}
			checkAnomalies(logtest.Scoped(t), db, clock, client)

			require.Equal(t, wantPayloads, client.payloads)
		})
	}
}
