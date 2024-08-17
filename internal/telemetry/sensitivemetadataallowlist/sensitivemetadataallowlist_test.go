package sensitivemetadataallowlist

import (
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/telemetry"
	v1 "github.com/sourcegraph/sourcegraph/lib/telemetrygateway/v1"
)

func TestIsAllowed(t *testing.T) {
	allowedTypes := AllowedEventTypes()
	require.NotEmpty(t, allowedTypes)

	for _, tc := range []struct {
		name            string
		event           *v1.Event
		expectAllowed   bool
		expectAllowlist []string
	}{
		{
			name: "allowed event",
			event: &v1.Event{
				Feature: string(telemetry.FeatureExample),
				Action:  string(telemetry.ActionExample),
			},
			expectAllowed:   true,
			expectAllowlist: []string{"testField"},
		},
		{
			name: "disallowed event",
			event: &v1.Event{
				Feature: "disallowedFeature",
				Action:  "disallowedAction",
			},
			expectAllowed: false,
		},
		{
			name: "event with private metadata, that has fields it is allowed to export (feature only)",
			// This feature's action is defined as a wildcard(*) in the "IsAllowed" check's list of known events
			event: &v1.Event{
				Feature: "cody.completion",
				Action:  "accepted",
			},
			expectAllowed: true,
			expectAllowlist: []string{
				"languageId",
				"otherCompletionProviders",
			},
		},
		{
			name: "event with private metadata, that has fields it is allowed to export (feature and action)",
			event: &v1.Event{
				Feature: "cody.hoverCommands",
				Action:  "visible",
			},
			expectAllowed: true,
			expectAllowlist: []string{
				"languageId",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			allowedKeys, allowed := allowedTypes.IsAllowed(tc.event)
			assert.Equal(t, tc.expectAllowed, allowed)
			assert.Equal(t, tc.expectAllowlist, allowedKeys)
		})
	}
}

func TestParseAdditionalAllowedEventTypes(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config string

		expect      autogold.Value
		expectError autogold.Value
	}{
		{
			name:        "invalid",
			config:      "asdf,foobar",
			expectError: autogold.Expect(`cannot parse SRC_TELEMETRY_SENSITIVEMETADATA_ADDITIONAL_ALLOWED_EVENT_TYPES value "asdf"`),
		},
		{
			name:        "invalid, no fields",
			config:      "foo::bar",
			expectError: autogold.Expect(`cannot parse SRC_TELEMETRY_SENSITIVEMETADATA_ADDITIONAL_ALLOWED_EVENT_TYPES value "foo::bar", missing allowlisted fields`),
		},
		{
			name:   "1 type",
			config: "foo::bar::field",
			expect: autogold.Expect([]EventType{{
				Feature:                    "foo",
				Action:                     "bar",
				AllowedPrivateMetadataKeys: []string{"field"},
			}}),
		},
		{
			name:   "multiple types",
			config: "foo::bar::field::field2,baz.bar::*::field",
			expect: autogold.Expect([]EventType{
				{
					Feature: "foo",
					Action:  "bar",
					AllowedPrivateMetadataKeys: []string{
						"field",
						"field2",
					},
				},
				{
					Feature:                    "baz.bar",
					Action:                     "*",
					AllowedPrivateMetadataKeys: []string{"field"},
				},
			}),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseAdditionalAllowedEventTypes(tc.config)
			if tc.expectError != nil {
				require.Error(t, err)
				tc.expectError.Equal(t, err.Error())
			} else {
				assert.NoError(t, err)
				tc.expect.Equal(t, got)
			}
		})
	}
}

func TestEventTypesRedact(t *testing.T) {
	allowedTypes := eventTypes(EventType{
		Feature:                    "example",
		Action:                     "exampleAction",
		AllowedPrivateMetadataKeys: []string{"foo"},
	})

	t.Run("dotcom mode", func(t *testing.T) {
		dotcom.MockSourcegraphDotComMode(t, true)
		mode := allowedTypes.Redact(&v1.Event{
			Feature: "example",
			Action:  "exampleAction",
		})
		assert.Equal(t, redactNothing, mode)

		ev := &v1.Event{
			Feature: "foobar",
			Action:  "exampleAction",
			Parameters: &v1.EventParameters{
				PrivateMetadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo": {
							Kind: &structpb.Value_NumberValue{
								NumberValue: 1,
							},
						},
					},
				},
			},
		}
		mode = allowedTypes.Redact(ev)
		assert.Equal(t, redactNothing, mode)
		assert.NotNil(t, ev.Parameters.PrivateMetadata)
	})

	t.Run("default", func(t *testing.T) {
		t.Run("allowlisted", func(t *testing.T) {
			mode := allowedTypes.Redact(&v1.Event{
				Feature: "example",
				Action:  "exampleAction",
			})
			assert.Equal(t, redactMarketingAndUnallowedPrivateMetadataKeys, mode)
		})
		t.Run("not allowlisted", func(t *testing.T) {
			ev := &v1.Event{
				Feature: "foobar",
				Action:  "exampleAction",
				Parameters: &v1.EventParameters{
					PrivateMetadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": {
								Kind: &structpb.Value_NumberValue{
									NumberValue: 1,
								},
							},
						},
					},
				},
			}
			mode := allowedTypes.Redact(ev)
			assert.Equal(t, redactAllSensitive, mode)
			assert.Nil(t, ev.Parameters.PrivateMetadata)
		})
		t.Run("allowlisted with wildcard(*) action", func(t *testing.T) {
			allowedTypes := eventTypes(EventType{
				Feature:                    "example",
				Action:                     "*",
				AllowedPrivateMetadataKeys: []string{"foo"},
			})
			ev := &v1.Event{
				Feature: "example",
				Action:  "randomAction",
				Parameters: &v1.EventParameters{
					PrivateMetadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo": {
								Kind: &structpb.Value_StringValue{
									StringValue: "allowed",
								},
							},
							"bar": {
								Kind: &structpb.Value_StringValue{
									StringValue: "redacted",
								},
							},
						},
					},
				}}
			mode := allowedTypes.Redact(ev)
			assert.Equal(t, redactMarketingAndUnallowedPrivateMetadataKeys, mode)

			// assert that only the allowlisted privateMetadata key (foo) has a value
			assert.Equal(t, "allowed", ev.Parameters.PrivateMetadata.Fields["foo"].GetStringValue())
			assert.Nil(t, ev.Parameters.PrivateMetadata.Fields["bar"])
		})
	})
}
