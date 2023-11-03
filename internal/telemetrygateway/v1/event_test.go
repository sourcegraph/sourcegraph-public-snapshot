package v1_test

import (
	context "context"
	"encoding/json"
	"testing"
	"time"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/tracetest"

	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
)

func TestNewEventWithDefaults(t *testing.T) {
	staticTime, err := time.Parse(time.RFC3339, "2023-02-24T14:48:30Z")
	require.NoError(t, err)

	t.Run("empty context", func(t *testing.T) {
		ctx := context.Background()
		got := telemetrygatewayv1.NewEventWithDefaults(ctx, staticTime, func() string { return "id" })
		assert.Nil(t, got.User)
		assert.Nil(t, got.Interaction)
		assert.Nil(t, got.FeatureFlags)
	})

	t.Run("extract actor and trace", func(t *testing.T) {
		// Atach a user
		var userID int32 = 123
		ctx := actor.WithActor(context.Background(), actor.FromMockUser(userID))

		// Attach a trace
		tracetest.ConfigureStaticTracerProvider(t)
		_, ctx = trace.New(ctx, t.Name())

		// NOTE: We can't test the feature flag part easily because
		// featureflag.GetEvaluatedFlagSet depends on Redis, and the package
		// is not designed for it to easily be stubbed out for testing.
		// Since it's used for existing telemetry, we trust it works.

		got := telemetrygatewayv1.NewEventWithDefaults(ctx, staticTime, func() string { return "id" })
		assert.NotNil(t, got.User)

		protodata, err := protojson.Marshal(got)
		require.NoError(t, err)

		// Protojson output isn't stable by injecting randomized whitespace,
		// so we re-marshal it to stabilize the output for golden tests.
		// https://github.com/golang/protobuf/issues/1082
		var gotJSON map[string]any
		require.NoError(t, json.Unmarshal(protodata, &gotJSON))
		jsondata, err := json.MarshalIndent(gotJSON, "", "  ")
		require.NoError(t, err)
		autogold.Expect(`{
  "id": "id",
  "interaction": {
    "traceId": "01020304050607080102040810203040"
  },
  "timestamp": "2023-02-24T14:48:30Z",
  "user": {
    "userId": "123"
  }
}`).Equal(t, string(jsondata))
	})
}
