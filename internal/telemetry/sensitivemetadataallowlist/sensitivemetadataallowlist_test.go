package sensitivemetadataallowlist

import (
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/telemetry"
	v1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
)

func TestIsAllowed(t *testing.T) {
	allowedTypes := AllowedEventTypes()
	require.NotEmpty(t, allowedTypes)
	assert.True(t, allowedTypes.IsAllowed(&v1.Event{
		Feature: string(telemetry.FeatureExample),
		Action:  string(telemetry.ActionExample),
	}))
	assert.False(t, allowedTypes.IsAllowed(&v1.Event{
		Feature: "disallowedFeature",
		Action:  "disallowedAction",
	}))
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
			name:   "1 type",
			config: "foo::bar",
			expect: autogold.Expect([]EventType{{
				Feature: "foo",
				Action:  "bar",
			}}),
		},
		{
			name:   "multiple types",
			config: "foo::bar,baz.bar::bar.baz",
			expect: autogold.Expect([]EventType{
				{
					Feature: "foo",
					Action:  "bar",
				},
				{
					Feature: "baz.bar",
					Action:  "bar.baz",
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
