package azureopenai

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type mockAzureClient struct {
	getCompletionsStream     func(ctx context.Context, body azopenai.CompletionsOptions, options *azopenai.GetCompletionsStreamOptions) (azopenai.GetCompletionsStreamResponse, error)
	getCompletions           func(ctx context.Context, body azopenai.CompletionsOptions, options *azopenai.GetCompletionsOptions) (azopenai.GetCompletionsResponse, error)
	getChatCompletions       func(ctx context.Context, body azopenai.ChatCompletionsOptions, options *azopenai.GetChatCompletionsOptions) (azopenai.GetChatCompletionsResponse, error)
	getChatCompletionsStream func(ctx context.Context, body azopenai.ChatCompletionsOptions, options *azopenai.GetChatCompletionsStreamOptions) (azopenai.GetChatCompletionsStreamResponse, error)
}

func (c *mockAzureClient) GetCompletionsStream(ctx context.Context, body azopenai.CompletionsOptions, options *azopenai.GetCompletionsStreamOptions) (azopenai.GetCompletionsStreamResponse, error) {
	return c.getCompletionsStream(ctx, body, options)
}

func (c *mockAzureClient) GetCompletions(ctx context.Context, body azopenai.CompletionsOptions, options *azopenai.GetCompletionsOptions) (azopenai.GetCompletionsResponse, error) {
	return c.getCompletions(ctx, body, options)
}

func (c *mockAzureClient) GetChatCompletions(ctx context.Context, body azopenai.ChatCompletionsOptions, options *azopenai.GetChatCompletionsOptions) (azopenai.GetChatCompletionsResponse, error) {
	return c.getChatCompletions(ctx, body, options)
}

func (c *mockAzureClient) GetChatCompletionsStream(ctx context.Context, body azopenai.ChatCompletionsOptions, options *azopenai.GetChatCompletionsStreamOptions) (azopenai.GetChatCompletionsStreamResponse, error) {
	return c.getChatCompletionsStream(ctx, body, options)
}

func getNewMockAzureAPIClient(mock *mockAzureClient) func(accessToken, endpoint string) (CompletionsClient, error) {
	return func(accessToken, endpoint string) (CompletionsClient, error) {
		return mock, nil
	}
}

func TestErrStatusNotOK(t *testing.T) {
	getAzureAPIClient := getNewMockAzureAPIClient(&mockAzureClient{
		getCompletionsStream: func(ctx context.Context, body azopenai.CompletionsOptions, options *azopenai.GetCompletionsStreamOptions) (azopenai.GetCompletionsStreamResponse, error) {
			return azopenai.GetCompletionsStreamResponse{}, errors.New("unexpected error")
		},
		getCompletions: func(ctx context.Context, body azopenai.CompletionsOptions, options *azopenai.GetCompletionsOptions) (azopenai.GetCompletionsResponse, error) {
			return azopenai.GetCompletionsResponse{}, errors.New("unexpected error")
		},
		getChatCompletions: func(ctx context.Context, body azopenai.ChatCompletionsOptions, options *azopenai.GetChatCompletionsOptions) (azopenai.GetChatCompletionsResponse, error) {
			return azopenai.GetChatCompletionsResponse{}, errors.New("unexpected error")
		},
		getChatCompletionsStream: func(ctx context.Context, body azopenai.ChatCompletionsOptions, options *azopenai.GetChatCompletionsStreamOptions) (azopenai.GetChatCompletionsStreamResponse, error) {
			return azopenai.GetChatCompletionsStreamResponse{}, errors.New("unexpected error")
		},
	})

	mockClient, _ := NewClient(getAzureAPIClient, "", "")
	t.Run("Complete", func(t *testing.T) {
		resp, err := mockClient.Complete(context.Background(), types.CompletionsFeatureChat, types.CompletionRequestParameters{})
		require.Error(t, err)
		assert.Nil(t, resp)

		autogold.Expect("unexpected error").Equal(t, err.Error())
		//_, ok := types.IsErrStatusNotOK(err)
		//assert.True(t, ok)
	})

	t.Run("Stream", func(t *testing.T) {
		err := mockClient.Stream(context.Background(), types.CompletionsFeatureChat, types.CompletionRequestParameters{}, func(event types.CompletionResponse) error { return nil })
		require.Error(t, err)

		autogold.Expect("unexpected error").Equal(t, err.Error())
		// _, ok := types.IsErrStatusNotOK(err)
		// assert.True(t, ok)
	})
}
