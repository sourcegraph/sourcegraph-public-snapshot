package openaicompatible

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

func NewClient(
	cli httpcli.Doer,
	modelConfigInfo *types.ModelConfigInfo,
	tokenManager tokenusage.Manager,
) types.CompletionsClient {
	return &client{
		cli:             cli,
		modelConfigInfo: modelConfigInfo,
		tokenManager:    tokenManager,
	}
}

type client struct {
	cli             httpcli.Doer
	modelConfigInfo *types.ModelConfigInfo
	tokenManager    tokenusage.Manager
}

func (c *client) Complete(
	ctx context.Context,
	logger log.Logger,
	request types.CompletionRequest,
) (*types.CompletionResponse, error) {
	var resp *http.Response
	var err error
	defer (func() {
		if resp != nil {
			resp.Body.Close()
		}
	})()

	if request.Feature == types.CompletionsFeatureCode {
		resp, err = c.makeCompletionRequest(ctx, request.Parameters, false)
	} else {
		resp, err = c.makeRequest(ctx, request.Parameters, false)
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response openaiResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	if len(response.Choices) == 0 {
		// Empty response.
		return &types.CompletionResponse{}, nil
	}
	modelID := c.modelConfigInfo.Model.ModelRef.ModelID()
	err = c.tokenManager.UpdateTokenCountsFromModelUsage(
		response.Usage.PromptTokens,
		response.Usage.CompletionTokens,
		tokenizer.OpenAIModel+"/"+string(modelID),
		string(request.Feature),
		tokenusage.OpenAICompatible)
	if err != nil {
		logger.Warn("Failed to count tokens with the token manager %w ", log.Error(err))
	}
	return &types.CompletionResponse{
		Completion: response.Choices[0].Text,
		StopReason: response.Choices[0].FinishReason,
	}, nil
}

func (c *client) Stream(
	ctx context.Context,
	logger log.Logger,
	request types.CompletionRequest,
	sendEvent types.SendCompletionEvent,
) error {
	var resp *http.Response
	var err error

	defer (func() {
		if resp != nil {
			resp.Body.Close()
		}
	})()
	if request.Feature == types.CompletionsFeatureCode {
		resp, err = c.makeCompletionRequest(ctx, request.Parameters, true)
	} else {
		resp, err = c.makeRequest(ctx, request.Parameters, true)
	}
	if err != nil {
		return err
	}

	dec := NewDecoder(resp.Body)
	var content string
	var ev types.CompletionResponse
	var promptTokens, completionTokens int

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
				content += event.Choices[0].Text
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
	modelID := c.modelConfigInfo.Model.ModelRef.ModelID()
	err = c.tokenManager.UpdateTokenCountsFromModelUsage(
		promptTokens,
		completionTokens,
		tokenizer.OpenAIModel+"/"+string(modelID),
		string(feature),
		tokenusage.OpenAICompatible,
	)
	if err != nil {
		logger.Warn("Failed to count tokens with the token manager %w", log.Error(err))
	}
	return nil
}

// makeRequest formats the request and calls the chat/completions endpoint for code_completion requests
func (c *client) makeRequest(
	ctx context.Context,
	requestParams types.CompletionRequestParameters,
	stream bool,
) (*http.Response, error) {
	if requestParams.TopK < 0 {
		requestParams.TopK = 0
	}
	if requestParams.TopP < 0 {
		requestParams.TopP = 0
	}

	// TODO(slimsag): self-hosted-models: there are other parameters here which we could allow
	// configuration of, which we do not specify as fields today:
	payload := openAIChatCompletionsRequestParameters{
		// TODO(slimsag): self-hosted-models: allow customization of model request param
		Model:       requestParams.Model,
		Temperature: requestParams.Temperature,
		TopP:        requestParams.TopP,
		// TODO(slimsag): self-hosted-models: allow customization of N (TopK?)
		N:         1,
		Stream:    stream,
		MaxTokens: requestParams.MaxTokensToSample,
		// TODO(slimsag): self-hosted-models: allow customization of stop sequences
		Stop: requestParams.StopSequences,
	}
	for _, m := range requestParams.Messages {
		// TODO(slimsag): map these 'roles' to openai system/user/assistant
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

	// TODO(slimsag): self-hosted-models: use OpenAICompatibleProvider
	endpoint := c.modelConfigInfo.Provider.ServerSideConfig.GenericProvider.Endpoint
	accessToken := c.modelConfigInfo.Provider.ServerSideConfig.GenericProvider.AccessToken
	url, err := url.Parse(endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse configured endpoint")
	}
	url.Path = "v1/chat/completions"

	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

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
func (c *client) makeCompletionRequest(
	ctx context.Context,
	requestParams types.CompletionRequestParameters,
	stream bool,
) (*http.Response, error) {
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
		// TODO(slimsag): self-hosted-models: allow customization of model request param
		Model:       requestParams.Model,
		Temperature: requestParams.Temperature,
		TopP:        requestParams.TopP,
		// TODO(slimsag): self-hosted-models: allow customization of N (TopK?)
		N:         1,
		Stream:    stream,
		MaxTokens: requestParams.MaxTokensToSample,
		// TODO(slimsag): self-hosted-models: allow customization of stop sequences
		Stop:   requestParams.StopSequences,
		Prompt: prompt,
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	// TODO(slimsag): self-hosted-models: use OpenAICompatibleProvider
	endpoint := c.modelConfigInfo.Provider.ServerSideConfig.GenericProvider.Endpoint
	accessToken := c.modelConfigInfo.Provider.ServerSideConfig.GenericProvider.AccessToken
	url, err := url.Parse(endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse configured endpoint")
	}
	url.Path = "v1/completions"

	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

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
