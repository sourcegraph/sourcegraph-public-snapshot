package sensitivemetadataallowlist

import (
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
			name: "disallowed event with additional allowed event type",
			event: &v1.Event{
				Feature: "cody.completions",
				Action:  "accepted",
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
				Feature: "foo",
				Action:  "bar",
				AllowedPrivateMetadataKeys: []string{
					"field",
				},
			}}),
		},
		{
			name:   "multiple types",
			config: "foo::bar::field::field2,baz.bar::bar.baz::field",
			expect: autogold.Expect([]EventType{
				{
					Feature: "foo",
					Action:  "bar",
					AllowedPrivateMetadataKeys: []string{
						"field", "field2",
					},
				},
				{
					Feature: "baz.bar",
					Action:  "bar.baz",
					AllowedPrivateMetadataKeys: []string{
						"field",
					},
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
