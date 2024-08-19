package openai

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/completions/tokenusage"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	modelconfigSDK "github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
)

type mockDoer struct {
	do func(*http.Request) (*http.Response, error)
}

func (c *mockDoer) Do(r *http.Request) (*http.Response, error) {
	return c.do(r)
}

var compRequest = types.CompletionRequest{
	Feature: types.CompletionsFeatureChat,
	Version: types.CompletionsVersionLegacy,
	ModelConfigInfo: types.ModelConfigInfo{
		Provider: modelconfigSDK.Provider{
			ID: modelconfigSDK.ProviderID("xxx-provider-id-xxx"),
		},
		Model: modelconfigSDK.Model{
			ModelRef: modelconfigSDK.ModelRef("provider::apiversion::test-model"),
		},
	},
	Parameters: types.CompletionRequestParameters{
		RequestedModel: "xxx-requested-model-xxx",
	},
}

func NewMockClient(statusCode int, response string) types.CompletionsClient {
	return NewClient(&mockDoer{
		func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: statusCode,
				Body:       io.NopCloser(bytes.NewReader([]byte(response))),
			}, nil
		},
	}, "", "", *tokenusage.NewManager())
}

func TestErrStatusNotOK(t *testing.T) {
	mockClient := NewMockClient(http.StatusTooManyRequests, "oh no, please slow down!")

	t.Run("Complete", func(t *testing.T) {
		logger := logtest.Scoped(t)
		resp, err := mockClient.Complete(context.Background(), logger, compRequest)
		require.Error(t, err)
		assert.Nil(t, resp)

		autogold.Expect("OpenAI: unexpected status code 429: oh no, please slow down!").Equal(t, err.Error())
		_, ok := types.IsErrStatusNotOK(err)
		assert.True(t, ok)
	})

	t.Run("Stream", func(t *testing.T) {
		logger := logtest.Scoped(t)
		sendEventFn := func(event types.CompletionResponse) error { return nil }
		err := mockClient.Stream(context.Background(), logger, compRequest, sendEventFn)
		require.Error(t, err)

		autogold.Expect("OpenAI: unexpected status code 429: oh no, please slow down!").Equal(t, err.Error())
		_, ok := types.IsErrStatusNotOK(err)
		assert.True(t, ok)
	})
}

func TestNonStreamingResponseParsing(t *testing.T) {
	mockClient := NewMockClient(http.StatusOK, `{
  "id": "chatcmpl-9wEJ9hnLdPcCLrfdZLrRPGOz48Pmo",
  "object": "chat.completion",
  "created": 1723665051,
  "model": "gpt-4o-mini-2024-07-18",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "yes",
        "refusal": null
      },
      "logprobs": null,
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 15,
    "completion_tokens": 1,
    "total_tokens": 16
  },
  "system_fingerprint": "fp_48196bc67a"
}`)
	logger := logtest.Scoped(t)
	resp, err := mockClient.Complete(context.Background(), logger, compRequest)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	autogold.Expect(&types.CompletionResponse{Completion: "yes", StopReason: "stop"}).Equal(t, resp)

}
