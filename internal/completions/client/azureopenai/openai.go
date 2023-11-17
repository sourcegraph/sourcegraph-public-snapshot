package azureopenai

import (
	"context"
	"io"
	"net/http"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// We want to reuse the client because when using the DefaultAzureCredential
// it will acquire a short lived token and reusing the client
// prevents acquiring a new token on every request.
// The client will refresh the token as needed.

var apiClient completionsClient

type completionsClient struct {
	mu          sync.RWMutex
	accessToken string
	endpoint    string
	client      *azopenai.Client
}

type CompletionsClient interface {
	GetCompletionsStream(ctx context.Context, body azopenai.CompletionsOptions, options *azopenai.GetCompletionsStreamOptions) (azopenai.GetCompletionsStreamResponse, error)
	GetCompletions(ctx context.Context, body azopenai.CompletionsOptions, options *azopenai.GetCompletionsOptions) (azopenai.GetCompletionsResponse, error)
	GetChatCompletions(ctx context.Context, body azopenai.ChatCompletionsOptions, options *azopenai.GetChatCompletionsOptions) (azopenai.GetChatCompletionsResponse, error)
	GetChatCompletionsStream(ctx context.Context, body azopenai.ChatCompletionsOptions, options *azopenai.GetChatCompletionsStreamOptions) (azopenai.GetChatCompletionsStreamResponse, error)
}

func GetAPIClient(endpoint, accessToken string) (CompletionsClient, error) {
	apiClient.mu.RLock()
	if apiClient.client != nil && apiClient.endpoint == endpoint && apiClient.accessToken == accessToken {
		apiClient.mu.RUnlock()
		return apiClient.client, nil
	}
	apiClient.mu.RUnlock()
	apiClient.mu.Lock()
	defer apiClient.mu.Unlock()
	var err error
	if accessToken != "" {
		credential, credErr := azopenai.NewKeyCredential(accessToken)
		if credErr != nil {
			return nil, credErr
		}
		apiClient.client, err = azopenai.NewClientWithKeyCredential(endpoint, credential, nil)
	} else {
		credential, credErr := azidentity.NewDefaultAzureCredential(nil)
		if credErr != nil {
			return nil, credErr
		}
		apiClient.client, err = azopenai.NewClient(endpoint, credential, nil)
	}
	return apiClient.client, err

}

type GetCompletionsAPIClientFunc func(accessToken, endpoint string) (CompletionsClient, error)

func NewClient(getClient GetCompletionsAPIClientFunc, accessToken, endpoint string) (types.CompletionsClient, error) {
	client, err := getClient(accessToken, endpoint)
	if err != nil {
		return nil, err
	}
	return &azureCompletionClient{
		client: client,
	}, nil
}

type azureCompletionClient struct {
	client CompletionsClient
}

func (c *azureCompletionClient) Complete(
	ctx context.Context,
	feature types.CompletionsFeature,
	requestParams types.CompletionRequestParameters,
) (*types.CompletionResponse, error) {

	switch feature {
	case types.CompletionsFeatureCode:
		return completeAutocomplete(ctx, c.client, feature, requestParams)
	case types.CompletionsFeatureChat:
		return completeChat(ctx, c.client, feature, requestParams)
	default:
		return nil, errors.New("invalid completions feature")
	}
}

func completeAutocomplete(
	ctx context.Context,
	client CompletionsClient,
	feature types.CompletionsFeature,
	requestParams types.CompletionRequestParameters,
) (*types.CompletionResponse, error) {
	options, err := getCompletionsOptions(requestParams)
	if err != nil {
		return nil, err
	}
	response, err := client.GetCompletions(ctx, options, nil)
	if err != nil {
		return nil, toStatusCodeError(err)
	}

	// Text and FinishReason are documented as REQUIRED but checking just to be safe
	if !hasValidFirstCompletionsChoice(response.Choices) {
		return &types.CompletionResponse{}, nil
	}
	return &types.CompletionResponse{
		Completion: *response.Choices[0].Text,
		StopReason: string(*response.Choices[0].FinishReason),
	}, nil
}

func completeChat(
	ctx context.Context,
	client CompletionsClient,
	feature types.CompletionsFeature,
	requestParams types.CompletionRequestParameters,
) (*types.CompletionResponse, error) {
	response, err := client.GetChatCompletions(ctx, getChatOptions(requestParams), nil)
	if err != nil {
		return nil, toStatusCodeError(err)
	}
	if !hasValidFirstChatChoice(response.Choices) {
		return &types.CompletionResponse{}, nil
	}
	return &types.CompletionResponse{
		Completion: *response.Choices[0].Delta.Content,
		StopReason: string(*response.Choices[0].FinishReason),
	}, nil
}

func (c *azureCompletionClient) Stream(
	ctx context.Context,
	feature types.CompletionsFeature,
	requestParams types.CompletionRequestParameters,
	sendEvent types.SendCompletionEvent,
) error {
	switch feature {
	case types.CompletionsFeatureCode:
		return streamAutocomplete(ctx, c.client, feature, requestParams, sendEvent)
	case types.CompletionsFeatureChat:
		return streamChat(ctx, c.client, feature, requestParams, sendEvent)
	default:
		return errors.New("invalid completions feature")
	}
}

func streamAutocomplete(
	ctx context.Context,
	client CompletionsClient,
	feature types.CompletionsFeature,
	requestParams types.CompletionRequestParameters,
	sendEvent types.SendCompletionEvent,
) error {
	options, err := getCompletionsOptions(requestParams)
	if err != nil {
		return err
	}
	resp, err := client.GetCompletionsStream(ctx, options, nil)
	if err != nil {
		return toStatusCodeError(err)
	}
	defer resp.CompletionsStream.Close()

	// Azure sends incremental deltas for each message in a chat stream
	// build up the full message content over multiple responses
	var content string
	for {
		entry, err := resp.CompletionsStream.Read()
		// stream is done
		if errors.Is(err, io.EOF) {
			return nil
		}
		// some other error has occured
		if err != nil {
			return err
		}

		// Text and Finish reason are marked as REQUIRED in documentation but check just in case
		if hasValidFirstCompletionsChoice(entry.Choices) {
			content += *entry.Choices[0].Text
			finish := string(*entry.Choices[0].FinishReason)
			ev := types.CompletionResponse{
				Completion: content,
				StopReason: finish,
			}
			return sendEvent(ev)
		}
	}
}

func streamChat(
	ctx context.Context,
	client CompletionsClient,
	feature types.CompletionsFeature,
	requestParams types.CompletionRequestParameters,
	sendEvent types.SendCompletionEvent,
) error {

	resp, err := client.GetChatCompletionsStream(ctx, getChatOptions(requestParams), nil)
	if err != nil {
		return toStatusCodeError(err)
	}
	defer resp.ChatCompletionsStream.Close()

	// Azure sends incremental deltas for each message in a chat stream
	// build up the full message content over multiple responses
	var content string
	for {
		entry, err := resp.ChatCompletionsStream.Read()
		// stream is done
		if errors.Is(err, io.EOF) {
			return nil
		}
		// some other error has occurred
		if err != nil {
			return err
		}

		if hasValidFirstChatChoice(entry.Choices) {
			// hasValidFirstChatChoice checks that FinishReason and Delta.Content aren't null
			// they are both marked as REQUIRED in docs despite being a pointer
			content += *entry.Choices[0].Delta.Content
			finish := string(*entry.Choices[0].FinishReason)
			ev := types.CompletionResponse{
				Completion: content,
				StopReason: finish,
			}
			err := sendEvent(ev)
			if err != nil {
				return err
			}
		}
	}
}

// hasValidChatChoice checks to ensure there is a choice and the first one contains non-nil values
func hasValidFirstChatChoice(choices []azopenai.ChatChoice) bool {
	return len(choices) > 0 &&
		choices[0].Delta != nil &&
		choices[0].Delta.Content != nil &&
		choices[0].FinishReason != nil
}

// hasValidChatChoice checks to ensure there is a choice and the first one contains non-nil values
func hasValidFirstCompletionsChoice(choices []azopenai.Choice) bool {
	return len(choices) > 0 &&
		choices[0].Text != nil &&
		choices[0].FinishReason != nil
}

func getChatMessages(messages []types.Message) []azopenai.ChatMessage {
	azureMessages := make([]azopenai.ChatMessage, len(messages))
	for i, m := range messages {
		var role azopenai.ChatRole
		message := m.Text
		switch m.Speaker {
		case types.HUMAN_MESSAGE_SPEAKER:
			role = azopenai.ChatRoleUser
		case types.ASISSTANT_MESSAGE_SPEAKER:
			role = azopenai.ChatRoleAssistant
		}
		azureMessages[i] = azopenai.ChatMessage{
			Content: &message,
			Role:    &role,
		}
	}
	return azureMessages
}

func getChatOptions(requestParams types.CompletionRequestParameters) azopenai.ChatCompletionsOptions {
	if requestParams.TopK < 0 {
		requestParams.TopK = 0
	}
	if requestParams.TopP < 0 {
		requestParams.TopP = 0
	}
	return azopenai.ChatCompletionsOptions{
		Messages:    getChatMessages(requestParams.Messages),
		Temperature: &requestParams.Temperature,
		TopP:        &requestParams.TopP,
		N:           intToInt32Ptr(1),
		Stop:        requestParams.StopSequences,
		MaxTokens:   intToInt32Ptr(requestParams.MaxTokensToSample),
		Deployment:  requestParams.Model,
	}
}

func getCompletionsOptions(requestParams types.CompletionRequestParameters) (azopenai.CompletionsOptions, error) {
	if requestParams.TopK < 0 {
		requestParams.TopK = 0
	}
	if requestParams.TopP < 0 {
		requestParams.TopP = 0
	}
	prompt, err := getPrompt(requestParams.Messages)
	if err != nil {
		return azopenai.CompletionsOptions{}, err
	}
	return azopenai.CompletionsOptions{
		Prompt:      []string{prompt},
		Temperature: &requestParams.Temperature,
		TopP:        &requestParams.TopP,
		N:           intToInt32Ptr(1),
		Stop:        requestParams.StopSequences,
		MaxTokens:   intToInt32Ptr(requestParams.MaxTokensToSample),
		Deployment:  requestParams.Model,
	}, nil
}

func getPrompt(messages []types.Message) (string, error) {
	if len(messages) != 1 {
		return "", errors.New("Expected to receive exactly one message with the prompt")
	}

	return messages[0].Text, nil
}

func intToInt32Ptr(i int) *int32 {
	v := int32(i)
	return &v
}

// toStatusCodeError converts Azure SDK ResponseError to a ErrStatusNotOK error
// when the status code is not OK.  This allows the request handler to return the
// appropriate status code to the calling client which is especially important for rate limits.
func toStatusCodeError(err error) error {
	var responseError *azcore.ResponseError
	if errors.As(err, &responseError) {
		if responseError.StatusCode != http.StatusOK {
			return types.NewErrStatusNotOK("AzureOpenAI", responseError.RawResponse)
		}
	}
	return err
}
