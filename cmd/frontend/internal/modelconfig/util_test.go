package modelconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/modelconfig/embedded"
	"github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
)

func TestDeepCopy(t *testing.T) {
	embeddedConfig, err := embedded.GetCodyGatewayModelConfig()
	require.NoError(t, err)

	copiedCfg, err := deepCopy(embeddedConfig)
	require.NoError(t, err)
	// Structural equality.
	assert.EqualValues(t, *embeddedConfig, *copiedCfg)
	// Referential equality.
	assert.True(t, embeddedConfig != copiedCfg)
}

func TestFilterListMatches(t *testing.T) {
	const geminiMRef = "google::v1::gemini-1.5-pro-latest"

	tests := []struct {
		MRef    string
		Pattern string
		Want    bool
	}{
		{
			MRef:    geminiMRef,
			Pattern: "google*",
			Want:    true,
		},
		{
			MRef:    geminiMRef,
			Pattern: "google::v1::*",
			Want:    true,
		},
		{
			MRef:    geminiMRef,
			Pattern: "*",
			Want:    true,
		},
		{
			MRef:    geminiMRef,
			Pattern: "*v1*",
			Want:    true,
		},
		{
			MRef:    geminiMRef,
			Pattern: "*::gemini-1.5-pro-latest",
			Want:    true,
		},
		{
			MRef:    geminiMRef,
			Pattern: "*::gpt-4o",
			Want:    false,
		},

		// Negative tests.
		{
			MRef:    geminiMRef,
			Pattern: "*::gpt-4o",
			Want:    false,
		},
		{
			MRef:    geminiMRef,
			Pattern: "*::v2::*",
			Want:    false,
		},
		{
			MRef:    geminiMRef,
			Pattern: "google::v1::gemini-1.5-pro", // Doesn't end with "-latest"
			Want:    false,
		},
		{
			MRef:    geminiMRef,
			Pattern: "anthropic*",
			Want:    false,
		},
	}

	for _, test := range tests {
		got := filterListMatches(types.ModelRef(test.MRef), []string{test.Pattern})
		assert.Equal(t, test.Want, got, "mref: %q\npattern: %q", test.MRef, test.Pattern)
	}
}
