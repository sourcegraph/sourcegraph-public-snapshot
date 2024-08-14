package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/completions/tokenizer"
	"github.com/sourcegraph/sourcegraph/internal/completions/tokenusage"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewClient(cli httpcli.Doer, endpoint, accessToken string, tokenManager tokenusage.Manager) types.CompletionsClient {
	return &openAIChatCompletionStreamClient{
		cli:          cli,
		accessToken:  accessToken,
		endpoint:     endpoint,
		tokenManager: tokenManager,
	}
}

type openAIChatCompletionStreamClient struct {
	cli          httpcli.Doer
	accessToken  string
	endpoint     string
	tokenManager tokenusage.Manager
}

func (c *openAIChatCompletionStreamClient) Complete(
	ctx context.Context,
	logger log.Logger,
	request types.CompletionRequest) (*types.CompletionResponse, error) {

	var (
		resp *http.Response
		err  error
	)
	defer (func() {
		if resp != nil {
			resp.Body.Close()
		}
	})()

	switch request.Feature {
	case types.CompletionsFeatureCode:
		resp, err = c.makeCompletionRequest(ctx, request, false)
	case types.CompletionsFeatureChat:
		resp, err = c.makeRequest(ctx, request, false)
	default:
		return nil, errors.Errorf("unknown feature %q", request.Feature)
	}
	if err != nil {
		return nil, err
	}

	var response openaiResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	if len(response.Choices) == 0 {
		logger.Warn("no choices in OpenAI response")
		return &types.CompletionResponse{}, nil
	}

	usage := response.Usage
	if err = c.recordTokenUsage(request, usage.PromptTokens, usage.CompletionTokens); err != nil {
		logger.Warn("Failed to count tokens with the token manager %w ", log.Error(err))
	}
	return &types.CompletionResponse{
		Completion: response.Choices[0].Message.Content,
		StopReason: response.Choices[0].FinishReason,
	}, nil
}

func (c *openAIChatCompletionStreamClient) Stream(
	ctx context.Context,
	logger log.Logger,
	request types.CompletionRequest,
	sendEvent types.SendCompletionEvent) error {

	var resp *http.Response
	var err error

	defer (func() {
		if resp != nil {
			resp.Body.Close()
		}
	})()
	switch request.Feature {
	case types.CompletionsFeatureCode:
		resp, err = c.makeCompletionRequest(ctx, request, true)
	case types.CompletionsFeatureChat:
		resp, err = c.makeRequest(ctx, request, true)
	default:
		return errors.Errorf("unknown feature %v", request.Feature)
	}
	if err != nil {
		return err
	}

	dec := NewDecoder(resp.Body)
	var (
		content                        string
		ev                             types.CompletionResponse
		promptTokens, completionTokens int
	)

	for dec.Scan() {
		if ctx.Err() != nil && ctx.Err() == context.Canceled {
			return nil
		}

		data := dec.Data()
		// Gracefully skip over any data that isn't JSON-like.
		if !bytes.HasPrefix(data, []byte("{")) {
			continue
		}

		var event openaiResponse
		if err := json.Unmarshal(data, &event); err != nil {
			return errors.Errorf("failed to decode event payload: %w - body: %s", err, string(data))
		}

		// These are only included in the last message, so we're not worried about overwriting
		if event.Usage.PromptTokens > 0 {
			promptTokens = event.Usage.PromptTokens
		}
		if event.Usage.CompletionTokens > 0 {
			completionTokens = event.Usage.CompletionTokens
		}

		if len(event.Choices) > 0 {
			if request.Feature == types.CompletionsFeatureCode {
				content += event.Choices[0].Message.Content
			} else {
				content += event.Choices[0].Delta.Content
			}
			ev = types.CompletionResponse{
				Completion: content,
				StopReason: event.Choices[0].FinishReason,
			}
			err = sendEvent(ev)
			if err != nil {
				return err
			}
		}
	}
	if dec.Err() != nil {
		return dec.Err()
	}

	if err = c.recordTokenUsage(request, promptTokens, completionTokens); err != nil {
		logger.Warn("Failed to count tokens with the token manager %w", log.Error(err))
	}
	return nil
}

func (c *openAIChatCompletionStreamClient) recordTokenUsage(request types.CompletionRequest, promptTokens, completionTokens int) error {
	feature := string(request.Feature)
	model := request.ModelConfigInfo.Model.ModelName
	label := tokenizer.OpenAIModel + "/" + string(model)
	return c.tokenManager.UpdateTokenCountsFromModelUsage(
		promptTokens, completionTokens,
		label, feature, tokenusage.OpenAI)
}

// makeRequest formats the request and calls the chat/completions endpoint for code_completion requests
func (c *openAIChatCompletionStreamClient) makeRequest(ctx context.Context, request types.CompletionRequest, stream bool) (*http.Response, error) {
	requestParams := request.Parameters
	if requestParams.TopK < 0 {
		requestParams.TopK = 0
	}
	if requestParams.TopP < 0 {
		requestParams.TopP = 0
	}

	// TODO(sqs): make CompletionRequestParameters non-anthropic-specific
	payload := openAIChatCompletionsRequestParameters{
		Model:       request.ModelConfigInfo.Model.ModelName,
		Temperature: requestParams.Temperature,
		TopP:        requestParams.TopP,
		// TODO(sqs): map requestParams.TopK to openai
		N:         1,
		Stream:    stream,
		MaxTokens: requestParams.MaxTokensToSample,
		// TODO: Our clients are currently heavily biased towards Anthropic,
		// so the stop sequences we send might not actually be very useful
		// for OpenAI.
		Stop: requestParams.StopSequences,
	}
	for _, m := range requestParams.Messages {
		// TODO(sqs): map these 'roles' to openai system/user/assistant
		var role string
		switch m.Speaker {
		case types.HUMAN_MESSAGE_SPEAKER:
			role = "user"
		case types.ASSISTANT_MESSAGE_SPEAKER:
			role = "assistant"
			//
		default:
			role = strings.ToLower(role)
		}
		payload.Messages = append(payload.Messages, message{
			Role:    role,
			Content: m.Text,
		})
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url, err := url.Parse(c.endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse configured endpoint")
	}
	url.Path = "v1/chat/completions"

	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, types.NewErrStatusNotOK("OpenAI", resp)
	}

	return resp, nil
}

// makeCompletionRequest formats the request and calls the completions endpoint for code_completion requests
func (c *openAIChatCompletionStreamClient) makeCompletionRequest(ctx context.Context, request types.CompletionRequest, stream bool) (*http.Response, error) {
	requestParams := request.Parameters
	if requestParams.TopK < 0 {
		requestParams.TopK = 0
	}
	if requestParams.TopP < 0 {
		requestParams.TopP = 0
	}

	prompt, err := getPrompt(requestParams.Messages)
	if err != nil {
		return nil, err
	}

	payload := openAICompletionsRequestParameters{
		Model:       request.ModelConfigInfo.Model.ModelName,
		Temperature: requestParams.Temperature,
		TopP:        requestParams.TopP,
		N:           1,
		Stream:      stream,
		MaxTokens:   requestParams.MaxTokensToSample,
		Stop:        requestParams.StopSequences,
		Prompt:      prompt,
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	url, err := url.Parse(c.endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse configured endpoint")
	}
	url.Path = "v1/completions"

	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, types.NewErrStatusNotOK("OpenAI", resp)
	}

	return resp, nil
}

func getPrompt(messages []types.Message) (string, error) {
	if l := len(messages); l != 1 {
		return "", errors.Errorf("expected to receive exactly one message with the prompt (got %d)", l)
	}

	return messages[0].Text, nil
}
