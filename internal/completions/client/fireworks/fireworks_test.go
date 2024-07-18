package fireworks

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
)

type mockDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (c *mockDoer) Do(r *http.Request) (*http.Response, error) {
	return c.do(r)
}

func TestErrStatusNotOK(t *testing.T) {
	mockClient := NewClient(&mockDoer{
		func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Body:       io.NopCloser(bytes.NewReader([]byte("oh no, please slow down!"))),
			}, nil
		},
	}, "", "")

	compRequest := types.CompletionRequest{
		Feature:         types.CompletionsFeatureCode,
		Version:         types.CompletionsVersionLegacy,
		ModelConfigInfo: types.ModelConfigInfo{
			// No provider or model information available.
			// We expect tests that need these values to fail.
		},
		Parameters: types.CompletionRequestParameters{
			Messages: []types.Message{
				{Text: "Hey"},
			},
		},
	}

	t.Run("Complete", func(t *testing.T) {
		logger := log.Scoped("completions")
		resp, err := mockClient.Complete(context.Background(), logger, compRequest)
		require.Error(t, err)
		assert.Nil(t, resp)

		autogold.Expect("Fireworks: unexpected status code 429: oh no, please slow down!").Equal(t, err.Error())
		_, ok := types.IsErrStatusNotOK(err)
		assert.True(t, ok)
	})

	t.Run("Stream", func(t *testing.T) {
		logger := log.Scoped("completions")
		sendEventFn := func(event types.CompletionResponse) error { return nil }
		err := mockClient.Stream(context.Background(), logger, compRequest, sendEventFn)
		require.Error(t, err)

		autogold.Expect("Fireworks: unexpected status code 429: oh no, please slow down!").Equal(t, err.Error())
		_, ok := types.IsErrStatusNotOK(err)
		assert.True(t, ok)
	})
}
