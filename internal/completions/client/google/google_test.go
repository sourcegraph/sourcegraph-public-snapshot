package google

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

	t.Run("Complete", func(t *testing.T) {
		logger := log.Scoped("completions")
		resp, err := mockClient.Complete(context.Background(), types.CompletionsFeatureChat, types.CompletionsVersionLegacy, types.CompletionRequestParameters{}, logger)
		require.Error(t, err)
		assert.Nil(t, resp)

		autogold.Expect("Google: unexpected status code 429: oh no, please slow down!").Equal(t, err.Error())
		_, ok := types.IsErrStatusNotOK(err)
		assert.True(t, ok)
	})

	t.Run("Stream", func(t *testing.T) {
		logger := log.Scoped("completions")
		err := mockClient.Stream(context.Background(), types.CompletionsFeatureChat, types.CompletionsVersionLegacy, types.CompletionRequestParameters{}, func(event types.CompletionResponse) error { return nil }, logger)
		require.Error(t, err)

		autogold.Expect("Google: unexpected status code 429: oh no, please slow down!").Equal(t, err.Error())
		_, ok := types.IsErrStatusNotOK(err)
		assert.True(t, ok)
	})
}

func TestGetAPIURL(t *testing.T) {
	t.Parallel()

	client := &googleCompletionStreamClient{
		endpoint:    "https://generativelanguage.googleapis.com/v1/models",
		accessToken: "test-token",
	}

	t.Run("valid v1 endpoint", func(t *testing.T) {
		params := types.CompletionRequestParameters{
			Model: "test-model",
		}
		url := client.getAPIURL(params, false).String()
		expected := "https://generativelanguage.googleapis.com/v1/models/test-model:generateContent?key=test-token"
		require.Equal(t, expected, url)
	})

	//
	t.Run("valid endpoint for Vertex AI", func(t *testing.T) {
		params := types.CompletionRequestParameters{
			Model: "gemini-1.5-pro",
		}
		c := &googleCompletionStreamClient{
			endpoint:    "https://vertex-ai.example.com/v1/projects/PROJECT_ID/locations/LOCATION/publishers/google/models",
			accessToken: "test-token",
		}
		url := c.getAPIURL(params, true).String()
		expected := "https://vertex-ai.example.com/v1/projects/PROJECT_ID/locations/LOCATION/publishers/google/models/gemini-1.5-pro:streamGenerateContent"
		require.Equal(t, expected, url)
	})

	t.Run("valid custom endpoint", func(t *testing.T) {
		params := types.CompletionRequestParameters{
			Model: "test-model",
		}
		c := &googleCompletionStreamClient{
			endpoint:    "https://example.com/api/models",
			accessToken: "test-token",
		}
		url := c.getAPIURL(params, true).String()
		expected := "https://example.com/api/models/test-model:streamGenerateContent"
		require.Equal(t, expected, url)
	})

	t.Run("invalid endpoint", func(t *testing.T) {
		client.endpoint = "://invalid"
		params := types.CompletionRequestParameters{
			Model: "test-model",
		}
		url := client.getAPIURL(params, false).String()
		expected := "https://generativelanguage.googleapis.com/v1beta/models/test-model:generateContent?key=test-token"
		require.Equal(t, expected, url)
	})

	t.Run("streaming", func(t *testing.T) {
		params := types.CompletionRequestParameters{
			Model: "test-model",
		}
		url := client.getAPIURL(params, true).String()
		expected := "https://generativelanguage.googleapis.com/v1beta/models/test-model:streamGenerateContent?alt=sse&key=test-token"
		require.Equal(t, expected, url)
	})

	t.Run("empty model", func(t *testing.T) {
		params := types.CompletionRequestParameters{
			Model: "",
		}
		url := client.getAPIURL(params, false).String()
		expected := "https://generativelanguage.googleapis.com/v1beta/models:generateContent?key=test-token"
		require.Equal(t, expected, url)
	})
}
