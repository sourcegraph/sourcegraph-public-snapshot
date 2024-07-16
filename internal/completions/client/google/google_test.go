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

	modelconfigSDK "github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
)

type mockDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (c *mockDoer) Do(r *http.Request) (*http.Response, error) {
	return c.do(r)
}

func TestErrStatusNotOK(t *testing.T) {
	mockClient, _ := NewClient(&mockDoer{
		func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Body:       io.NopCloser(bytes.NewReader([]byte("oh no, please slow down!"))),
			}, nil
		},
	}, "https://generativelanguage.googleapis.com", "", false)

	compRequest := types.CompletionRequest{
		Feature:    types.CompletionsFeatureChat,
		Version:    types.CompletionsVersionLegacy,
		Parameters: types.CompletionRequestParameters{},
	}

	t.Run("Complete", func(t *testing.T) {
		logger := log.Scoped("completions")
		resp, err := mockClient.Complete(context.Background(), logger, compRequest)
		require.Error(t, err)
		assert.Nil(t, resp)

		autogold.Expect("Google: unexpected status code 429: oh no, please slow down!").Equal(t, err.Error())
		_, ok := types.IsErrStatusNotOK(err)
		assert.True(t, ok)
	})

	t.Run("Stream", func(t *testing.T) {
		logger := log.Scoped("completions")
		sendEventFn := func(event types.CompletionResponse) error { return nil }
		err := mockClient.Stream(context.Background(), logger, compRequest, sendEventFn)
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

	buildRequestForModel := func(model string) types.CompletionRequest {
		// Note that this isn't a "valid" request, in that most metadata is missing.
		// So if we add more unit tests, we'll probably need to flesh this out more.
		return types.CompletionRequest{
			ModelConfigInfo: types.ModelConfigInfo{
				Model: modelconfigSDK.Model{
					ModelName: model,
				},
			},
		}
	}

	t.Run("valid v1 endpoint", func(t *testing.T) {
		req := buildRequestForModel("test-model")
		url := client.getAPIURL(req, false).String()
		expected := "https://generativelanguage.googleapis.com/v1/models/test-model:generateContent?key=test-token"
		require.Equal(t, expected, url)
	})

	t.Run("valid endpoint for Vertex AI", func(t *testing.T) {
		req := buildRequestForModel("gemini-1.5-pro")
		c := &googleCompletionStreamClient{
			endpoint:    "https://vertex-ai.example.com/v1/projects/PROJECT_ID/locations/LOCATION/publishers/google/models",
			accessToken: "test-token",
		}
		url := c.getAPIURL(req, true).String()
		expected := "https://vertex-ai.example.com/v1/projects/PROJECT_ID/locations/LOCATION/publishers/google/models/gemini-1.5-pro:streamGenerateContent"
		require.Equal(t, expected, url)
	})

	t.Run("valid custom endpoint", func(t *testing.T) {
		req := buildRequestForModel("test-model")
		c := &googleCompletionStreamClient{
			endpoint:    "https://example.com/api/models",
			accessToken: "test-token",
		}
		url := c.getAPIURL(req, true).String()
		expected := "https://example.com/api/models/test-model:streamGenerateContent"
		require.Equal(t, expected, url)
	})

	t.Run("invalid endpoint", func(t *testing.T) {
		client.endpoint = "://invalid"
		req := buildRequestForModel("test-model")
		url := client.getAPIURL(req, false).String()
		expected := "https://generativelanguage.googleapis.com/v1beta/models/test-model:generateContent?key=test-token"
		require.Equal(t, expected, url)
	})

	t.Run("streaming", func(t *testing.T) {
		req := buildRequestForModel("test-model")
		url := client.getAPIURL(req, true).String()
		expected := "https://generativelanguage.googleapis.com/v1beta/models/test-model:streamGenerateContent?alt=sse&key=test-token"
		require.Equal(t, expected, url)
	})

	t.Run("empty model", func(t *testing.T) {
		req := buildRequestForModel("")
		url := client.getAPIURL(req, false).String()
		expected := "https://generativelanguage.googleapis.com/v1beta/models:generateContent?key=test-token"
		require.Equal(t, expected, url)
	})
}
