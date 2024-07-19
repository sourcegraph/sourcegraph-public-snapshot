package openaicompatible

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/sourcegraph/log"

	sse "github.com/tmaxmax/go-sse"

	"github.com/sourcegraph/sourcegraph/internal/completions/tokenizer"
	"github.com/sourcegraph/sourcegraph/internal/completions/tokenusage"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	modelconfigSDK "github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewClient(
	cli *http.Client,
	tokenManager tokenusage.Manager,
) types.CompletionsClient {
	return &client{
		cli:          cli,
		tokenManager: tokenManager,
		rng:          rand.New(rand.NewSource(time.Now().Unix())),
	}
}

type client struct {
	cli          *http.Client
	tokenManager tokenusage.Manager
	rng          *rand.Rand
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
		req, reqBody, err = c.makeCompletionRequest(ctx, request, false)
	} else {
		req, reqBody, err = c.makeChatRequest(ctx, request, false)
	}
	if err != nil {
		return nil, errors.Wrap(err, "making request")
	}

	requestID := c.rng.Uint32()
	providerConfig := request.ModelConfigInfo.Provider.ServerSideConfig.OpenAICompatible
	if providerConfig.EnableVerboseLogs {
		logger.Info("request",
			log.Uint32("id", requestID),
			log.String("kind", "non-streaming"),
			log.String("method", req.Method),
			log.String("url", req.URL.String()),
			// Note: log package will automatically redact token
			log.String("headers", fmt.Sprint(req.Header)),
			log.String("body", reqBody),
		)
	}
	start := time.Now()
	resp, err = c.cli.Do(req)
	if err != nil {
		logger.Error("request error",
			log.Uint32("id", requestID),
			log.Error(err),
		)
		return nil, errors.Wrap(err, "performing request")
	}
	if resp.StatusCode != http.StatusOK {
		err := types.NewErrStatusNotOK("OpenAI", resp)
		logger.Error("request error",
			log.Uint32("id", requestID),
			log.Error(err),
		)
		return nil, err
	}
	defer resp.Body.Close()

	var response openaiResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		logger.Error("request error, decoding response",
			log.Uint32("id", requestID),
			log.Error(err),
		)
		return nil, errors.Wrap(err, "decoding response")
	}
	if providerConfig.EnableVerboseLogs {
		// When debugging connections, log more verbose information like the actual completion we got back.
		completion := ""
		if len(response.Choices) > 0 {
			completion = response.Choices[0].Text
		}
		logger.Info("request success",
			log.Uint32("id", requestID),
			log.Duration("time", time.Since(start)),
			log.String("response_model", response.Model),
			log.String("url", req.URL.String()),
			log.String("system_fingerprint", response.SystemFingerprint),
			log.String("finish_reason", response.maybeGetFinishReason()),
			log.String("completion", completion),
		)
	} else {
		logger.Info("request success",
			log.Uint32("id", requestID),
			log.Duration("time", time.Since(start)),
			log.String("response_model", response.Model),
			log.String("url", req.URL.String()),
			log.String("system_fingerprint", response.SystemFingerprint),
			log.String("finish_reason", response.maybeGetFinishReason()),
		)
	}

	if len(response.Choices) == 0 {
		// Empty response.
		return &types.CompletionResponse{}, nil
	}

	modelID := request.ModelConfigInfo.Model.ModelRef.ModelID()
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
		req, reqBody, err = c.makeCompletionRequest(ctx, request, true)
	} else {
		req, reqBody, err = c.makeChatRequest(ctx, request, true)
	}
	if err != nil {
		return errors.Wrap(err, "making request")
	}

	sseClient := &sse.Client{
		HTTPClient:        c.cli,
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

	var (
		content                        string
		ev                             types.CompletionResponse
		promptTokens, completionTokens int
		streamErr                      error
		finishReason                   string
	)
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
					finishReason = choice.FinishReason
					streamErr = nil
					cancel()
					return
				}
			}
		}
	})
	defer unsubscribe()

	requestID := c.rng.Uint32()
	providerConfig := request.ModelConfigInfo.Provider.ServerSideConfig.OpenAICompatible
	if providerConfig.EnableVerboseLogs {
		logger.Info("request",
			log.Uint32("id", requestID),
			log.String("kind", "streaming"),
			log.String("method", req.Method),
			log.String("url", req.URL.String()),
			// Note: log package will automatically redact token
			log.String("headers", fmt.Sprint(req.Header)),
			log.String("body", reqBody),
		)
	}
	start := time.Now()
	err = conn.Connect()
	if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
		// go-sse will return io.EOF on successful close of the connection, since it expects the
		// connection to be long-lived. In our case, we expect the connection to close on success
		// and be short lived, so this is a non-error.
		err = nil
	}
	if streamErr != nil {
		err = errors.Append(err, streamErr)
	}
	if err == nil && finishReason == "" {
		// At this point, we successfully streamed the response to the client. But we need to make
		// sure the client gets a non-empty StopReason at the very end, otherwise it would think
		// the streamed response it got is partial / incomplete and may not display the completion
		// to the user as a result.
		err = sendEvent(types.CompletionResponse{
			Completion: content,
			StopReason: "stop_sequence", // pretend we hit a stop sequence (we did!)
		})
	}
	if err != nil {
		logger.Error("request error",
			log.Uint32("id", requestID),
			log.Error(err),
		)
		return errors.Wrap(err, "NewConnection")
	}

	if providerConfig.EnableVerboseLogs {
		// When debugging connections, log more verbose information like the actual completion we got back.
		logger.Info("request success",
			log.Uint32("id", requestID),
			log.Duration("time", time.Since(start)),
			log.String("url", req.URL.String()),
			log.String("finish_reason", finishReason),
			log.String("completion", content),
		)
	} else {
		logger.Info("request success",
			log.Uint32("id", requestID),
			log.Duration("time", time.Since(start)),
			log.String("url", req.URL.String()),
			log.String("finish_reason", finishReason),
		)
	}

	modelID := request.ModelConfigInfo.Model.ModelRef.ModelID()
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

func (c *client) makeChatRequest(
	ctx context.Context,
	request types.CompletionRequest,
	stream bool,
) (*http.Request, string, error) {
	requestParams := request.Parameters
	if requestParams.TopK < 0 {
		requestParams.TopK = 0
	}
	if requestParams.TopP < 0 {
		requestParams.TopP = 0
	}

	payload := openAIChatCompletionsRequestParameters{
		Model:       getAPIModel(request),
		Temperature: requestParams.Temperature,
		TopP:        requestParams.TopP,
		N:           requestParams.TopK,
		Stream:      stream,
		MaxTokens:   requestParams.MaxTokensToSample,
		Stop:        requestParams.StopSequences,
	}
	for _, m := range requestParams.Messages {
		var role string
		switch m.Speaker {
		case types.SYSTEM_MESSAGE_SPEAKER:
			role = "system"
		case types.HUMAN_MESSAGE_SPEAKER:
			role = "user"
		case types.ASSISTANT_MESSAGE_SPEAKER:
			role = "assistant"
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
		return nil, "", errors.Wrap(err, "Marshal")
	}

	endpoint, err := getEndpoint(request, c.rng)
	if err != nil {
		return nil, "", errors.Wrap(err, "getEndpoint")
	}
	url, err := getEndpointURL(endpoint, "chat/completions")
	if err != nil {
		return nil, "", errors.Wrap(err, "getEndpointURL")
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, "", errors.Wrap(err, "NewRequestWithContext")
	}

	req.Header.Set("Content-Type", "application/json")
	if endpoint.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+endpoint.AccessToken)
	}
	return req, string(reqBody), nil
}

func (c *client) makeCompletionRequest(
	ctx context.Context,
	request types.CompletionRequest,
	stream bool,
) (*http.Request, string, error) {
	requestParams := request.Parameters
	if requestParams.TopK < 0 {
		requestParams.TopK = 0
	}
	if requestParams.TopP < 0 {
		requestParams.TopP = 0
	}

	prompt, err := getPrompt(requestParams.Messages)
	if err != nil {
		return nil, "", errors.Wrap(err, "getPrompt")
	}

	payload := openAICompletionsRequestParameters{
		Model:       getAPIModel(request),
		Temperature: requestParams.Temperature,
		TopP:        requestParams.TopP,
		N:           requestParams.TopK,
		Stream:      stream,
		MaxTokens:   requestParams.MaxTokensToSample,
		Stop:        requestParams.StopSequences,
		Prompt:      prompt,
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, "", errors.Wrap(err, "Marshal")
	}

	endpoint, err := getEndpoint(request, c.rng)
	if err != nil {
		return nil, "", errors.Wrap(err, "getEndpoint")
	}
	url, err := getEndpointURL(endpoint, "completions")
	if err != nil {
		return nil, "", errors.Wrap(err, "getEndpointURL")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, "", errors.Wrap(err, "NewRequestWithContext")
	}

	req.Header.Set("Content-Type", "application/json")
	if endpoint.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+endpoint.AccessToken)
	}
	return req, string(reqBody), nil
}

func getPrompt(messages []types.Message) (string, error) {
	if l := len(messages); l == 0 {
		return "", errors.New("found zero messages in prompt")
	}
	return messages[0].Text, nil
}

func getAPIModel(request types.CompletionRequest) string {
	ssConfig := request.ModelConfigInfo.Model.ServerSideConfig
	if ssConfig != nil && ssConfig.OpenAICompatible != nil && ssConfig.OpenAICompatible.APIModel != "" {
		return ssConfig.OpenAICompatible.APIModel
	}
	// Default to model name if not specified
	return request.ModelConfigInfo.Model.ModelName
}

func getEndpoint(request types.CompletionRequest, rng *rand.Rand) (modelconfigSDK.OpenAICompatibleEndpoint, error) {
	providerConfig := request.ModelConfigInfo.Provider.ServerSideConfig.OpenAICompatible
	if len(providerConfig.Endpoints) == 0 {
		return modelconfigSDK.OpenAICompatibleEndpoint{}, errors.New("no openaicompatible endpoint configured")
	}
	if len(providerConfig.Endpoints) == 1 {
		return providerConfig.Endpoints[0], nil
	}
	randPick := rng.Intn(len(providerConfig.Endpoints))
	return providerConfig.Endpoints[randPick], nil
}

func getEndpointURL(endpoint modelconfigSDK.OpenAICompatibleEndpoint, relativePath string) (*url.URL, error) {
	url, err := url.Parse(endpoint.URL)
	if err != nil {
		return nil, errors.Newf("failed to parse endpoint URL: %q", endpoint.URL)
	}
	if url.Scheme == "" || url.Host == "" {
		return nil, errors.Newf("unable to build URL, bad endpoint: %q", endpoint.URL)
	}
	url.Path = path.Join(url.Path, relativePath)
	return url, nil
}
