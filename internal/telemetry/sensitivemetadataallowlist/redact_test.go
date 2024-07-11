package sensitivemetadataallowlist

import (
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/lib/telemetrygateway/v1"
)

func TestRedactEvent(t *testing.T) {
	makeFullEvent := func() *telemetrygatewayv1.Event {
		return &telemetrygatewayv1.Event{
			Parameters: &telemetrygatewayv1.EventParameters{
				PrivateMetadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"testField":    structpb.NewStringValue("TestValue"),
						"notTestField": structpb.NewStringValue("notTestValue"),
						"nestedTestField": structpb.NewStructValue(&structpb.Struct{
							Fields: map[string]*structpb.Value{
								"testField":    structpb.NewStringValue("TestValue"),
								"notTestField": structpb.NewStringValue("notTestValue"),
							},
						}),
						"boolTestField": structpb.NewBoolValue(true),
					},
				},
			},
			MarketingTracking: &telemetrygatewayv1.EventMarketingTracking{
				Url: pointers.Ptr("sourcegraph.com"),
			},
		}
	}
	tests := []struct {
		name        string
		mode        redactMode
		event       *telemetrygatewayv1.Event
		allowedKeys []string
		assert      func(t *testing.T, got *telemetrygatewayv1.Event)
	}{
		{
			name:  "redact all sensitive",
			mode:  redactAllSensitive,
			event: makeFullEvent(),
			assert: func(t *testing.T, got *telemetrygatewayv1.Event) {
				assert.Nil(t, got.Parameters.PrivateMetadata)
				assert.Nil(t, got.MarketingTracking)
			},
		},
		{
			name:  "redact all sensitive on empty event",
			mode:  redactAllSensitive,
			event: &telemetrygatewayv1.Event{},
			assert: func(t *testing.T, got *telemetrygatewayv1.Event) {
				assert.Nil(t, got.Parameters.PrivateMetadata)
				assert.Nil(t, got.MarketingTracking)
			},
		},
		{
			name:  "redact marketing",
			mode:  redactMarketingAndUnallowedPrivateMetadataKeys,
			event: makeFullEvent(),
			allowedKeys: []string{
				"testField",
			},
			assert: func(t *testing.T, got *telemetrygatewayv1.Event) {
				assert.Nil(t, got.MarketingTracking)
				require.NotNil(t, got.Parameters.PrivateMetadata)
				assert.NotNil(t, got.Parameters.PrivateMetadata.Fields["testField"])
				assert.Nil(t, got.Parameters.PrivateMetadata.Fields["notTestField"])
			},
		},
		{
			name:  "redact non-string type on allowlist",
			mode:  redactMarketingAndUnallowedPrivateMetadataKeys,
			event: makeFullEvent(),
			allowedKeys: []string{
				"boolTestField",
				"nestedTestField",
			},
			assert: func(t *testing.T, got *telemetrygatewayv1.Event) {
				// check that non-string types are redacted, only string types are allowed on allowlist
				autogold.Expect("ERROR: value of allowlisted key was not a string, got: *structpb.Value_BoolValue").Equal(t, got.Parameters.PrivateMetadata.Fields["boolTestField"].GetStringValue())
				autogold.Expect("ERROR: value of allowlisted key was not a string, got: *structpb.Value_StructValue").Equal(t, got.Parameters.PrivateMetadata.Fields["nestedTestField"].GetStringValue())
			},
		},
		{
			name:  "redact nothing",
			mode:  redactNothing,
			event: makeFullEvent(),
			assert: func(t *testing.T, got *telemetrygatewayv1.Event) {
				assert.NotNil(t, got.Parameters.PrivateMetadata)
				assert.NotNil(t, got.MarketingTracking)
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ev := makeFullEvent()
			redactEvent(ev, tc.mode, tc.allowedKeys)
			tc.assert(t, ev)
		})
	}
}
