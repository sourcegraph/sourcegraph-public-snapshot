package sensitivemetadataallowlist

import (
	"testing"

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
