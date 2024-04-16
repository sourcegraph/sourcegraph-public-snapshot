package resolvers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

func TestNewTelemetryGatewayEvents(t *testing.T) {
	staticTime, err := time.Parse(time.RFC3339, "2023-02-24T14:48:30Z")
	require.NoError(t, err)

	for _, tc := range []struct {
		name   string
		ctx    context.Context
		event  graphqlbackend.TelemetryEventInput
		expect autogold.Value
	}{
		{
			name: "basic",
			ctx:  context.Background(),
			event: graphqlbackend.TelemetryEventInput{
				Feature: "Feature",
				Action:  "Example",
			},
			expect: autogold.Expect(`{
  "action": "Example",
  "feature": "Feature",
  "id": "basic",
  "parameters": {},
  "source": {
    "client": {},
    "server": {
      "version": "0.0.0+dev"
    }
  },
  "timestamp": "2023-02-24T14:48:30Z"
}`),
		},
		{
			name: "with anonymous user",
			ctx:  actor.WithActor(context.Background(), actor.FromAnonymousUser("1234")),
			event: graphqlbackend.TelemetryEventInput{
				Feature: "Feature",
				Action:  "Example",
			},
			expect: autogold.Expect(`{
  "action": "Example",
  "feature": "Feature",
  "id": "with anonymous user",
  "parameters": {},
  "source": {
    "client": {},
    "server": {
      "version": "0.0.0+dev"
    }
  },
  "timestamp": "2023-02-24T14:48:30Z",
  "user": {
    "anonymousUserId": "1234"
  }
}`),
		},
		{
			name: "with authenticated user",
			ctx:  actor.WithActor(context.Background(), actor.FromMockUser(1234)),
			event: graphqlbackend.TelemetryEventInput{
				Feature: "Feature",
				Action:  "Example",
			},
			expect: autogold.Expect(`{
  "action": "Example",
  "feature": "Feature",
  "id": "with authenticated user",
  "parameters": {},
  "source": {
    "client": {},
    "server": {
      "version": "0.0.0+dev"
    }
  },
  "timestamp": "2023-02-24T14:48:30Z",
  "user": {
    "userId": "1234"
  }
}`),
		},
		{
			name: "with parameters",
			ctx:  context.Background(),
			event: graphqlbackend.TelemetryEventInput{
				Feature: "Feature",
				Action:  "Example",
				Parameters: graphqlbackend.TelemetryEventParametersInput{
					Version: 0,
					Metadata: &[]graphqlbackend.TelemetryEventMetadataInput{
						{
							Key:   "metadata",
							Value: graphqlbackend.JSONValue{Value: 123},
						},
					},
					PrivateMetadata: &graphqlbackend.JSONValue{
						Value: map[string]any{"private": "super-sensitive"},
					},
					BillingMetadata: &graphqlbackend.TelemetryEventBillingMetadataInput{
						Product:  "Product",
						Category: "Category",
					},
				},
			},
			expect: autogold.Expect(`{
  "action": "Example",
  "feature": "Feature",
  "id": "with parameters",
  "parameters": {
    "billingMetadata": {
      "category": "Category",
      "product": "Product"
    },
    "metadata": {
      "metadata": 123
    },
    "privateMetadata": {
      "private": "super-sensitive"
    }
  },
  "source": {
    "client": {},
    "server": {
      "version": "0.0.0+dev"
    }
  },
  "timestamp": "2023-02-24T14:48:30Z"
}`),
		},
		{
			name: "with string PrivateMetadata",
			ctx:  context.Background(),
			event: graphqlbackend.TelemetryEventInput{
				Feature: "Feature",
				Action:  "Example",
				Parameters: graphqlbackend.TelemetryEventParametersInput{
					Version: 0,
					PrivateMetadata: &graphqlbackend.JSONValue{
						Value: "some metadata",
					},
				},
			},
			expect: autogold.Expect(`{
  "action": "Example",
  "feature": "Feature",
  "id": "with string PrivateMetadata",
  "parameters": {
    "privateMetadata": {
      "value": "some metadata"
    }
  },
  "source": {
    "client": {},
    "server": {
      "version": "0.0.0+dev"
    }
  },
  "timestamp": "2023-02-24T14:48:30Z"
}`),
		},
		{
			name: "with numeric PrivateMetadata",
			ctx:  context.Background(),
			event: graphqlbackend.TelemetryEventInput{
				Feature: "Feature",
				Action:  "Example",
				Parameters: graphqlbackend.TelemetryEventParametersInput{
					Version: 0,
					PrivateMetadata: &graphqlbackend.JSONValue{
						Value: 1234,
					},
				},
			},
			expect: autogold.Expect(`{
  "action": "Example",
  "feature": "Feature",
  "id": "with numeric PrivateMetadata",
  "parameters": {
    "privateMetadata": {
      "value": 1234
    }
  },
  "source": {
    "client": {},
    "server": {
      "version": "0.0.0+dev"
    }
  },
  "timestamp": "2023-02-24T14:48:30Z"
}`),
		},
		{
			name: "with custom timestamp",
			ctx:  context.Background(),
			event: graphqlbackend.TelemetryEventInput{
				Timestamp: &gqlutil.DateTime{Time: staticTime.Add(48 * time.Hour)},
				Feature:   "Feature",
				Action:    "Example",
			},
			expect: autogold.Expect(`{
  "action": "Example",
  "feature": "Feature",
  "id": "with custom timestamp",
  "parameters": {},
  "source": {
    "client": {},
    "server": {
      "version": "0.0.0+dev"
    }
  },
  "timestamp": "2023-02-26T14:48:30Z"
}`),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := newTelemetryGatewayEvents(tc.ctx,
				staticTime,
				func() string { return tc.name },
				[]graphqlbackend.TelemetryEventInput{
					tc.event,
				})
			require.NoError(t, err)
			require.Len(t, got, 1)

			protodata, err := protojson.Marshal(got[0])
			require.NoError(t, err)

			// Protojson output isn't stable by injecting randomized whitespace,
			// so we re-marshal it to stabilize the output for golden tests.
			// https://github.com/golang/protobuf/issues/1082
			var gotJSON map[string]any
			require.NoError(t, json.Unmarshal(protodata, &gotJSON))
			jsondata, err := json.MarshalIndent(gotJSON, "", "  ")
			require.NoError(t, err)
			tc.expect.Equal(t, string(jsondata))
		})
	}
}
