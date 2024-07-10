package openaicompatible

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/sourcegraph/log"

	sse "github.com/tmaxmax/go-sse"

	"github.com/sourcegraph/sourcegraph/internal/completions/tokenizer"
	"github.com/sourcegraph/sourcegraph/internal/completions/tokenusage"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// TODO(slimsag): self-hosted-models: expose this as an option
const debugConnections = true

func NewClient(
	cli *http.Client,
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
	cli             *http.Client
	modelConfigInfo *types.ModelConfigInfo
	tokenManager    tokenusage.Manager
}

func (c *client) Complete(
	ctx context.Context,
	logger log.Logger,
	request types.CompletionRequest,
) (*types.CompletionResponse, error) {
	logger = logger.Scoped("OpenAICompatible")

	var resp *http.Response
	defer (func() {
		if resp != nil {
			resp.Body.Close()
		}
	})()

	var (
		req     *http.Request
		reqBody string
		err     error
	)
	if request.Feature == types.CompletionsFeatureCode {
		req, reqBody, err = c.makeCompletionRequest(ctx, request.Parameters, false)
	} else {
		req, reqBody, err = c.makeChatRequest(ctx, request.Parameters, false)
	}
	if err != nil {
		return nil, errors.Wrap(err, "making request")
	}

	if debugConnections {
		logger.Info("request",
			log.String("kind", "non-streaming"),
			log.String("method", req.Method),
			log.String("url", req.URL.String()),
			// Note: log package will automatically redact token
			log.String("headers", fmt.Sprint(req.Header)),
			log.String("body", reqBody),
		)
	}
	resp, err = c.cli.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "performing request")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, types.NewErrStatusNotOK("OpenAI", resp)
	}
	defer resp.Body.Close()

	var response openaiResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "decoding response")
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
	logger = logger.Scoped("OpenAICompatible")

	var (
		req     *http.Request
		reqBody string
		err     error
	)
	if request.Feature == types.CompletionsFeatureCode {
		req, reqBody, err = c.makeCompletionRequest(ctx, request.Parameters, true)
	} else {
		req, reqBody, err = c.makeChatRequest(ctx, request.Parameters, true)
	}
	if err != nil {
		return errors.Wrap(err, "making request")
	}

	sseClient := &sse.Client{
		HTTPClient: c.cli,
		// TODO(slimsag): self-hosted-models: investigate non-default validator implications
		ResponseValidator: sse.DefaultValidator,
		Backoff: sse.Backoff{
			// Note: go-sse has a bug with retry logic (https://github.com/tmaxmax/go-sse/pull/38)
			// where it will get stuck in an infinite retry loop due to an io.EOF error
			// depending on how the server behaves. For now, we just do not expose retry/backoff
			// logic. It's not really useful for these types of requests anyway given their
			// short-lived nature.
			MaxRetries: -1,
		},
	}
	ctx, cancel := context.WithCancel(ctx)
	conn := sseClient.NewConnection(req.WithContext(ctx))

	var content string
	var ev types.CompletionResponse
	var promptTokens, completionTokens int
	var streamErr error

	unsubscribe := conn.SubscribeMessages(func(event sse.Event) {
		// Ignore any data that is not JSON-like
		if !strings.HasPrefix(event.Data, "{") {
			return
		}

		var resp openaiResponse
		if err := json.Unmarshal([]byte(event.Data), &resp); err != nil {
			streamErr = errors.Errorf("failed to decode event payload: %w - body: %s", err, event.Data)
			cancel()
			return
		}

		if reflect.DeepEqual(resp, openaiResponse{}) {
			// Empty response, it may be an error payload then
			var errResp openaiErrorResponse
			if err := json.Unmarshal([]byte(event.Data), &errResp); err != nil {
				streamErr = errors.Errorf("failed to decode error event payload: %w - body: %s", err, event.Data)
				cancel()
				return
			}
			if errResp.Error != "" || errResp.ErrorType != "" {
				streamErr = errors.Errorf("SSE error: %s: %s", errResp.ErrorType, errResp.Error)
				cancel()
				return
			}
		}

		// These are only included in the last message, so we're not worried about overwriting
		if resp.Usage.PromptTokens > 0 {
			promptTokens = resp.Usage.PromptTokens
		}
		if resp.Usage.CompletionTokens > 0 {
			completionTokens = resp.Usage.CompletionTokens
		}

		if len(resp.Choices) > 0 {
			if request.Feature == types.CompletionsFeatureCode {
				content += resp.Choices[0].Text
			} else {
				content += resp.Choices[0].Delta.Content
			}
			ev = types.CompletionResponse{
				Completion: content,
				StopReason: resp.Choices[0].FinishReason,
			}
			err = sendEvent(ev)
			if err != nil {
				streamErr = errors.Errorf("failed to send event: %w", err)
				cancel()
				return
			}
			for _, choice := range resp.Choices {
				if choice.FinishReason != "" {
					// End of stream
					streamErr = nil
					cancel()
					return
				}
			}
		}
	})
	defer unsubscribe()

	if debugConnections {
		logger.Info("request",
			log.String("kind", "streaming"),
			log.String("method", req.Method),
			log.String("url", req.URL.String()),
			// Note: log package will automatically redact token
			log.String("headers", fmt.Sprint(req.Header)),
			log.String("body", reqBody),
		)
	}
	err = conn.Connect()
	if (err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, io.EOF)) || streamErr != nil {
		err = errors.Append(err, streamErr)
		logger.Info("request error", log.Error(err))
		return errors.Wrap(err, "NewConnection")
	}
	if debugConnections {
		logger.Info("request success")
	}

	modelID := c.modelConfigInfo.Model.ModelRef.ModelID()
	err = c.tokenManager.UpdateTokenCountsFromModelUsage(
		promptTokens,
		completionTokens,
		tokenizer.OpenAIModel+"/"+string(modelID),
		string(request.Feature),
		tokenusage.OpenAICompatible,
	)
	if err != nil {
		logger.Warn("Failed to count tokens with the token manager %w", log.Error(err))
	}
	return nil
}

// TODO(slimsag): self-hosted-models: wrap errors below this point in the file

func (c *client) makeChatRequest(
	ctx context.Context,
	requestParams types.CompletionRequestParameters,
	stream bool,
) (*http.Request, string, error) {
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
		return nil, "", err
	}

	// TODO(slimsag): self-hosted-models: use OpenAICompatibleProvider
	endpoint := c.modelConfigInfo.Provider.ServerSideConfig.GenericProvider.Endpoint
	url, err := url.Parse(endpoint)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to parse configured endpoint")
	}
	url.Path = "v1/chat/completions"
	if url.Scheme == "" || url.Host == "" {
		return nil, "", errors.Newf("unable to build URL, bad endpoint: %q", endpoint)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, "", err
	}

	// TODO(slimsag): self-hosted-models: allow setting custom headers
	req.Header.Set("Content-Type", "application/json")
	// TODO(slimsag): self-hosted-models: use OpenAICompatibleProvider
	accessToken := c.modelConfigInfo.Provider.ServerSideConfig.GenericProvider.AccessToken
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	return req, string(reqBody), nil
}

func (c *client) makeCompletionRequest(
	ctx context.Context,
	requestParams types.CompletionRequestParameters,
	stream bool,
) (*http.Request, string, error) {
	if requestParams.TopK < 0 {
		requestParams.TopK = 0
	}
	if requestParams.TopP < 0 {
		requestParams.TopP = 0
	}

	prompt, err := getPrompt(requestParams.Messages)
	if err != nil {
		return nil, "", err
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
		return nil, "", err
	}
	// TODO(slimsag): self-hosted-models: use OpenAICompatibleProvider
	endpoint := c.modelConfigInfo.Provider.ServerSideConfig.GenericProvider.Endpoint
	url, err := url.Parse(endpoint)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to parse configured endpoint")
	}
	url.Path = "v1/completions"
	if url.Scheme == "" || url.Host == "" {
		return nil, "", errors.Newf("unable to build URL, bad endpoint: %q", endpoint)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, "", err
	}

	req.Header.Set("Content-Type", "application/json")
	// TODO(slimsag): self-hosted-models: use OpenAICompatibleProvider
	accessToken := c.modelConfigInfo.Provider.ServerSideConfig.GenericProvider.AccessToken
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	return req, string(reqBody), nil
}

func getPrompt(messages []types.Message) (string, error) {
	if l := len(messages); l != 1 {
		// TODO(slimsag): self-hosted-models: relax oft-problematic constraint
		return "", errors.Errorf("expected to receive exactly one message with the prompt (got %d)", l)
	}
	return messages[0].Text, nil
}
