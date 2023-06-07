package codygateway

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProviderFromGatewayModel(t *testing.T) {
	for _, tc := range []struct {
		gatewayModel string

		expectProvider string
		expectModel    string
	}{
		{gatewayModel: "anthropic/claude-v1",
			expectProvider: "anthropic", expectModel: "claude-v1"},
		{gatewayModel: "openai/gpt4",
			expectProvider: "openai", expectModel: "gpt4"},

		// Edge cases
		{gatewayModel: "claude-v1",
			expectProvider: "", expectModel: "claude-v1"},
		{gatewayModel: "openai/unexpectednamewith/slash",
			expectProvider: "openai", expectModel: "unexpectednamewith/slash"},
	} {
		t.Run(tc.gatewayModel, func(t *testing.T) {
			p, m := getProviderFromGatewayModel(tc.gatewayModel)
			assert.Equal(t, tc.expectProvider, p)
			assert.Equal(t, tc.expectModel, m)
		})
	}
}
