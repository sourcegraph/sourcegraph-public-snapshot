package teestore

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// see TestRecorderEndToEnd for tests that include teestore.Store and the database

func TestToEventLogs(t *testing.T) {
	testCases := []struct {
		name            string
		events          []*telemetrygatewayv1.Event
		expectEventLogs autogold.Value
	}{
		{
			name:            "handles all nil",
			events:          nil,
			expectEventLogs: autogold.Expect("[]"),
		},
		{
			name:   "handles nil entry",
			events: []*telemetrygatewayv1.Event{nil},
			expectEventLogs: autogold.Expect(`[
  {
    "ID": 0,
    "Name": ".",
    "URL": "",
    "UserID": 0,
    "AnonymousUserID": "",
    "Argument": null,
    "PublicArgument": null,
    "Source": "BACKEND",
    "Version": "",
    "Timestamp": "2022-11-03T02:00:00Z",
    "EvaluatedFlagSet": {},
    "CohortID": null,
    "FirstSourceURL": null,
    "LastSourceURL": null,
    "Referrer": null,
    "DeviceID": null,
    "InsertID": null,
    "Client": null,
    "BillingProductCategory": null,
    "BillingEventID": null
  }
]`),
		},
		{
			name:   "handles nil fields",
			events: []*telemetrygatewayv1.Event{{}},
			expectEventLogs: autogold.Expect(`[
  {
    "ID": 0,
    "Name": ".",
    "URL": "",
    "UserID": 0,
    "AnonymousUserID": "",
    "Argument": null,
    "PublicArgument": null,
    "Source": "BACKEND",
    "Version": "",
    "Timestamp": "2022-11-03T02:00:00Z",
    "EvaluatedFlagSet": {},
    "CohortID": null,
    "FirstSourceURL": null,
    "LastSourceURL": null,
    "Referrer": null,
    "DeviceID": null,
    "InsertID": null,
    "Client": null,
    "BillingProductCategory": null,
    "BillingEventID": null
  }
]`),
		},
		{
			name: "simple event",
			events: []*telemetrygatewayv1.Event{{
				Id:        "1",
				Timestamp: timestamppb.New(time.Date(2022, 11, 2, 1, 0, 0, 0, time.UTC)),
				Feature:   "CodeSearch",
				Action:    "Seach",
				Source: &telemetrygatewayv1.EventSource{
					Client: &telemetrygatewayv1.EventSource_Client{
						Name:    "VSCODE",
						Version: pointers.Ptr("1.2.3"),
					},
					Server: &telemetrygatewayv1.EventSource_Server{
						Version: "dev",
					},
				},
				Parameters: &telemetrygatewayv1.EventParameters{
					Metadata: map[string]int64{"public": 2},
					PrivateMetadata: &structpb.Struct{Fields: map[string]*structpb.Value{
						"private": structpb.NewStringValue("sensitive-data"),
					}},
					BillingMetadata: &telemetrygatewayv1.EventBillingMetadata{
						Product:  "product",
						Category: "category",
					},
				},
				User: &telemetrygatewayv1.EventUser{
					UserId:          pointers.Ptr(int64(1234)),
					AnonymousUserId: pointers.Ptr("anonymous"),
				},
				MarketingTracking: &telemetrygatewayv1.EventMarketingTracking{
					Url: pointers.Ptr("sourcegraph.com/foobar"),
				},
			}},
			expectEventLogs: autogold.Expect(`[
  {
    "ID": 0,
    "Name": "VSCODE:CodeSearch.Seach",
    "URL": "sourcegraph.com/foobar",
    "UserID": 1234,
    "AnonymousUserID": "anonymous",
    "Argument": {
      "private": "sensitive-data"
    },
    "PublicArgument": {
      "public": 2
    },
    "Source": "VSCODE",
    "Version": "dev",
    "Timestamp": "2022-11-02T01:00:00Z",
    "EvaluatedFlagSet": {},
    "CohortID": null,
    "FirstSourceURL": null,
    "LastSourceURL": null,
    "Referrer": null,
    "DeviceID": null,
    "InsertID": null,
    "Client": "VSCODE:1.2.3",
    "BillingProductCategory": "category",
    "BillingEventID": null
  }
]`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			eventLogs := toEventLogs(
				func() time.Time { return time.Date(2022, 11, 3, 2, 0, 0, 0, time.UTC) },
				tc.events)
			require.Len(t, eventLogs, len(tc.events))
			// Compare JSON for ease of reading
			data, err := json.MarshalIndent(eventLogs, "", "  ")
			require.NoError(t, err)
			tc.expectEventLogs.Equal(t, string(data))
		})
	}
}
