package cloud

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSiteConfig(t *testing.T) {
	const testRawSiteConfig = "eyJzaWduYXR1cmUiOnsiRm9ybWF0Ijoic3NoLWVkMjU1MTkiLCJCbG9iIjoiMlFVdVNvUTNsZEdieVpXb280MTMxbEN0YTRtWlBWNm9MTENVMEVqWHlpYm4zK0hhZzMrSkwrdzd5Ulk1RlorN2pRZFR0Szg2Vk8wVGhYUVJWcTZyQUE9PSIsIlJlc3QiOm51bGx9LCJzaXRlQ29uZmlnIjoiZXdvZ0lDSmhkWFJvVUhKdmRtbGtaWEp6SWpvZ2V3b2dJQ0FnSW5OdmRYSmpaV2R5WVhCb1QzQmxjbUYwYjNJaU9pQjdDaUFnSUNBZ0lDSnBjM04xWlhJaU9pQWlhSFIwY0hNNkx5OWtaWFl0ZEdWemRDNXZhM1JoY0hKbGRtbGxkeTVqYjIwaUxBb2dJQ0FnSUNBaVkyeHBaVzUwU1VRaU9pQWlkR1Z6ZEVOc2FXVnVkRWxFSWl3S0lDQWdJQ0FnSW1Oc2FXVnVkRk5sWTNKbGRDSTZJQ0owWlhOMFEyeHBaVzUwVTJWamNtVjBJaXdLSUNBZ0lDQWdJbXhwWm1WamVXTnNaVVIxY21GMGFXOXVJam9nTVRBS0lDQWdJSDBLSUNCOUNuMEsifQ"
	got, err := parseSiteConfig(testRawSiteConfig)
	require.NoError(t, err)

	want := &SchemaSiteConfig{
		AuthProviders: &SchemaAuthProviders{
			SourcegraphOperator: &SchemaAuthProviderSourcegraphOperator{
				Issuer:            "https://dev-test.oktapreview.com",
				ClientID:          "testClientID",
				ClientSecret:      "testClientSecret",
				LifecycleDuration: 10,
			},
		},
	}
	assert.Equal(t, want, got)
}
