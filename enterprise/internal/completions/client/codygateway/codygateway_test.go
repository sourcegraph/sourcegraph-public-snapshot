package codygateway

import (
	"net/http/httptest"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

func TestOverwriteErrorSource(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.WriteHeader(500)
	originalErr := types.NewErrStatusNotOK("Foobar", rec.Result())

	err := overwriteErrSource(originalErr)
	require.Error(t, err)
	statusErr, ok := types.IsErrStatusNotOK(err)
	require.True(t, ok)
	autogold.Expect("Sourcegraph Cody Gateway").Equal(t, statusErr.Source)

	assert.NoError(t, overwriteErrSource(nil))
	assert.Equal(t, "asdf", overwriteErrSource(errors.New("asdf")).Error())
}
