package azureopenai

import (
	"context"
	"io"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// We want to reuse the client because when using the DefaultAzureCredential
// it will acquire a short lived token and reusing the client
// prevents acquiring a new token on every request.
// The client will refresh the token as needed.
var azureClient *azopenai.Client

type AzureCompletionsClient interface {
	GetCompletionsStream(ctx context.Context, body azopenai.CompletionsOptions, options *azopenai.GetCompletionsStreamOptions) (azopenai.GetCompletionsStreamResponse, error)
	GetCompletions(ctx context.Context, body azopenai.CompletionsOptions, options *azopenai.GetCompletionsOptions) (azopenai.GetCompletionsResponse, error)
	GetChatCompletions(ctx context.Context, body azopenai.ChatCompletionsOptions, options *azopenai.GetChatCompletionsOptions) (azopenai.GetChatCompletionsResponse, error)
	GetChatCompletionsStream(ctx context.Context, body azopenai.ChatCompletionsOptions, options *azopenai.GetChatCompletionsStreamOptions) (azopenai.GetChatCompletionsStreamResponse, error)
}

func GetAzureAPIClient(endpoint, accessToken string) (AzureCompletionsClient, error) {
	if azureClient != nil {
		return azureClient, nil
	}
	var err error
	if accessToken != "" {
		credential, credErr := azopenai.NewKeyCredential(accessToken)
		if credErr != nil {
			return nil, credErr
		}
		azureClient, err = azopenai.NewClientWithKeyCredential(endpoint, credential, nil)
	} else {
		credential, credErr := azidentity.NewDefaultAzureCredential(nil)
		if credErr != nil {
			return nil, credErr
		}
		azureClient, err = azopenai.NewClient(endpoint, credential, nil)
	}
	return azureClient, err
}

func NewClient(client AzureCompletionsClient) (types.CompletionsClient, error) {
	return &azureCompletionClient{
		client: client,
	}, nil
}

type azureCompletionClient struct {
	client AzureCompletionsClient
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
	client AzureCompletionsClient,
	feature types.CompletionsFeature,
	requestParams types.CompletionRequestParameters,
) (*types.CompletionResponse, error) {
	options, err := getCompletionsOptions(requestParams)
	if err != nil {
		return nil, err
	}
	response, err := client.GetCompletions(ctx, options, nil)
	if err != nil {
		return nil, err
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
	client AzureCompletionsClient,
	feature types.CompletionsFeature,
	requestParams types.CompletionRequestParameters,
) (*types.CompletionResponse, error) {
	response, err := client.GetChatCompletions(ctx, getChatOptions(requestParams), nil)
	if err != nil {
		return nil, err
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
	client AzureCompletionsClient,
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
		return err
	}
	defer resp.CompletionsStream.Close()

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
			ev := types.CompletionResponse{
				Completion: *entry.Choices[0].Text,
				StopReason: string(*entry.Choices[0].FinishReason),
			}
			return sendEvent(ev)
		}
	}
}

func streamChat(
	ctx context.Context,
	client AzureCompletionsClient,
	feature types.CompletionsFeature,
	requestParams types.CompletionRequestParameters,
	sendEvent types.SendCompletionEvent,
) error {

	resp, err := client.GetChatCompletionsStream(ctx, getChatOptions(requestParams), nil)
	if err != nil {
		return err
	}
	defer resp.ChatCompletionsStream.Close()

	for {
		entry, err := resp.ChatCompletionsStream.Read()
		// stream is done
		if errors.Is(err, io.EOF) {
			return nil
		}
		// some other error has occured
		if err != nil {
			return err
		}

		// Delta and Finish reason are marked as REQUIRED in documentation but check just in case
		if hasValidFirstChatChoice(entry.Choices) {
			ev := types.CompletionResponse{
				Completion: *entry.Choices[0].Delta.Content,
				StopReason: string(*entry.Choices[0].FinishReason),
			}
			return sendEvent(ev)
		}
	}
}

// hasValidChatChoice checks to ensure there is a choice and the first one contains non-nil values
func hasValidFirstChatChoice(choices []azopenai.ChatChoice) bool {
	return len(choices) > 0 &&
		choices[0].Delta != nil &&
		choices[0].FinishReason != nil &&
		choices[0].Delta.Content != nil
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
		switch m.Speaker {
		case types.HUMAN_MESSAGE_SPEAKER:
			role = azopenai.ChatRoleUser
		case types.ASISSTANT_MESSAGE_SPEAKER:
			role = azopenai.ChatRoleAssistant
		}
		azureMessages[i] = azopenai.ChatMessage{
			Content: &m.Text,
			Role:    &role,
		}
	}
	return azureMessages
}

func getChatOptions(requestParams types.CompletionRequestParameters) azopenai.ChatCompletionsOptions {
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
