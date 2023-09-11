package telemetry

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/sourcegraph/sourcegraph/internal/actor"
)

func TestMakeRawEvent(t *testing.T) {
	staticTime, err := time.Parse(time.RFC3339, "2023-02-24T14:48:30Z")
	require.NoError(t, err)

	for _, tc := range []struct {
		name   string
		ctx    context.Context
		event  Event
		expect autogold.Value
	}{
		{
			name: "basic",
			ctx:  context.Background(),
			event: Event{
				Name: EventExample,
			},
			expect: autogold.Expect(`{
"name": "Example",
"parameters": {},
"source": {
	"serverVersion": "0.0.0+dev"
},
"timestamp": "2023-02-24T14:48:30Z",
"user": {}
}`),
		},
		{
			name: "with user",
			ctx:  actor.WithActor(context.Background(), actor.FromAnonymousUser("1234")),
			event: Event{
				Name: EventExample,
			},
			expect: autogold.Expect(`{
"name": "Example",
"parameters": {},
"source": {
	"serverVersion": "0.0.0+dev"
},
"timestamp": "2023-02-24T14:48:30Z",
"user": {
	"anonymousUserID": "1234"
}
}`),
		},
		{
			name: "with parameters",
			ctx:  context.Background(),
			event: Event{
				Name: EventExample,
				Parameters: EventParameters{
					Version: 0,
					Metadata: EventMetadata{
						"foobar": 3,
					},
					PrivateMetadata: map[string]any{
						"barbaz": "hello world!",
					},
					BillingMetadata: &EventBillingMetadata{
						Product:  BillingProductExample,
						Category: BillingCategoryExample,
					},
				},
			},
			expect: autogold.Expect(`{
"name": "Example",
"parameters": {
	"billingMetadata": {
		"category": 5,
		"product": 3
	},
	"metadata": {
		"foobar": 3
	},
	"privateMetadata": "{\"barbaz\":\"hello world!\"}"
},
"source": {
	"serverVersion": "0.0.0+dev"
},
"timestamp": "2023-02-24T14:48:30Z",
"user": {}
}`),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := makeRawEvent(tc.ctx, staticTime, tc.event.Name, tc.event.Parameters)

			protodata, err := protojson.Marshal(&got)
			require.NoError(t, err)

			// Protojson output isn't stable by injecting randomized whitespace,
			// so we re-marshal it to stabilize the output for golden tests.
			// https://github.com/golang/protobuf/issues/1082
			var gotJSON map[string]any
			require.NoError(t, json.Unmarshal(protodata, &gotJSON))
			jsondata, err := json.MarshalIndent(gotJSON, "", "\t")
			require.NoError(t, err)
			tc.expect.Equal(t, string(jsondata))
		})
	}
}
