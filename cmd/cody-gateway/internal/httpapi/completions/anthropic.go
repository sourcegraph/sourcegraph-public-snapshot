package completions

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/shared/config"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/anthropic"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/notify"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/completions/tokenizer"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewAnthropicHandler(baseLogger log.Logger, eventLogger events.Logger, rs limiter.RedisStore, rateLimitNotifier notify.RateLimitNotifier, httpClient httpcli.Doer, config config.AnthropicConfig, promptRecorder PromptRecorder, upstreamConfig UpstreamHandlerConfig) (http.Handler, error) {
	// Tokenizer only needs to be initialized once, and can be shared globally.
	anthropicTokenizer, err := tokenizer.NewCL100kBaseTokenizer()
	if err != nil {
		return nil, err
	}

	// Anthropic primarily uses concurrent requests to rate-limit spikes
	// in requests, so set a default retry-after that is likely to be
	// acceptable for Sourcegraph clients to retry (the default
	// SRC_HTTP_CLI_EXTERNAL_RETRY_AFTER_MAX_DURATION) since we might be
	// able to circumvent concurrents limits without raising an error to the
	// user.
	upstreamConfig.DefaultRetryAfterSeconds = 2

	return makeUpstreamHandler[anthropicRequest](
		baseLogger,
		eventLogger,
		rs,
		rateLimitNotifier,
		httpClient,
		string(conftypes.CompletionsProviderNameAnthropic),
		config.AllowedModels,
		&AnthropicHandlerMethods{config: config, anthropicTokenizer: anthropicTokenizer},
		promptRecorder,
		upstreamConfig,
	), nil
}

// anthropicRequest captures all known fields from https://console.anthropic.com/docs/api/reference.
type anthropicRequest struct {
	Prompt            string                    `json:"prompt"`
	Model             string                    `json:"model"`
	MaxTokensToSample int32                     `json:"max_tokens_to_sample"`
	StopSequences     []string                  `json:"stop_sequences,omitempty"`
	Stream            bool                      `json:"stream,omitempty"`
	Temperature       float32                   `json:"temperature,omitempty"`
	TopK              int32                     `json:"top_k,omitempty"`
	TopP              float32                   `json:"top_p,omitempty"`
	Metadata          *anthropicRequestMetadata `json:"metadata,omitempty"`

	// Use (*anthropicRequest).GetTokenCount()
	promptTokens *anthropicTokenCount
}

func (ar anthropicRequest) ShouldStream() bool {
	return ar.Stream
}

func (ar anthropicRequest) GetModel() string {
	return ar.Model
}

func (ar anthropicRequest) BuildPrompt() string {
	return ar.Prompt
}

type anthropicTokenCount struct {
	count int
	err   error
}

// GetPromptTokenCount computes the token count of the prompt exactly once using
// the given tokenizer. It is not concurrency-safe.
func (ar *anthropicRequest) GetPromptTokenCount(tk tokenizer.Tokenizer) (int, error) {
	if ar.promptTokens == nil {
		tokens, err := tk.Tokenize(ar.Prompt)
		ar.promptTokens = &anthropicTokenCount{
			count: len(tokens),
			err:   err,
		}
	}
	return ar.promptTokens.count, ar.promptTokens.err
}

type anthropicRequestMetadata struct {
	UserID string `json:"user_id,omitempty"`
}

// anthropicResponse captures all relevant-to-us fields from https://console.anthropic.com/docs/api/reference.
type anthropicResponse struct {
	Completion string `json:"completion,omitempty"`
	StopReason string `json:"stop_reason,omitempty"`
}

type AnthropicHandlerMethods struct {
	anthropicTokenizer tokenizer.Tokenizer
	promptRecorder     PromptRecorder
	config             config.AnthropicConfig
}

func (a *AnthropicHandlerMethods) getAPIURL(feature codygateway.Feature, _ anthropicRequest) string {
	return "https://api.anthropic.com/v1/complete"
}

func (a *AnthropicHandlerMethods) validateRequest(ctx context.Context, logger log.Logger, _ codygateway.Feature, ar anthropicRequest) error {
	maxTokensToSample := a.config.FlaggingConfig.MaxTokensToSample
	if ar.MaxTokensToSample > int32(maxTokensToSample) {
		return errors.Errorf("max_tokens_to_sample exceeds maximum allowed value of %d: %d", maxTokensToSample, ar.MaxTokensToSample)
	}
	return nil
}

func (a *AnthropicHandlerMethods) shouldFlagRequest(ctx context.Context, logger log.Logger, ar anthropicRequest) (*flaggingResult, error) {
	result, err := isFlaggedRequest(a.anthropicTokenizer,
		flaggingRequest{
			ModelName:       ar.Model,
			FlattenedPrompt: ar.Prompt,
			MaxTokens:       int(ar.MaxTokensToSample),
		},
		makeFlaggingConfig(a.config.FlaggingConfig),
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *AnthropicHandlerMethods) transformBody(body *anthropicRequest, identifier string) {
	// Overwrite the metadata field, we don't want to allow users to specify it:
	body.Metadata = &anthropicRequestMetadata{
		// We forward the actor ID to support tracking.
		UserID: identifier,
	}
}

func (a *AnthropicHandlerMethods) getRequestMetadata(body anthropicRequest) (model string, additionalMetadata map[string]any) {
	return body.Model, map[string]any{
		"stream":               body.Stream,
		"max_tokens_to_sample": body.MaxTokensToSample,
	}
}

func (a *AnthropicHandlerMethods) transformRequest(downstreamRequest, upstreamRequest *http.Request) {
	// Mimic headers set by the official Anthropic client:
	// https://sourcegraph.com/github.com/anthropics/anthropic-sdk-typescript@493075d70f50f1568a276ed0cb177e297f5fef9f/-/blob/src/index.ts
	upstreamRequest.Header.Set("Cache-Control", "no-cache")
	upstreamRequest.Header.Set("Accept", "application/json")
	upstreamRequest.Header.Set("Content-Type", "application/json")
	upstreamRequest.Header.Set("Client", "sourcegraph-cody-gateway/1.0")
	upstreamRequest.Header.Set("X-API-Key", a.config.AccessToken)
	upstreamRequest.Header.Set("anthropic-version", "2023-01-01")
}

func (a *AnthropicHandlerMethods) parseResponseAndUsage(logger log.Logger, reqBody anthropicRequest, r io.Reader, isStreamRequest bool) (promptUsage, completionUsage usageStats) {
	var err error

	// Setting a default -1 value so that in case of errors the tokenizer computed tokens don't impact the data
	completionUsage.tokenizerTokens = -1
	promptUsage.tokenizerTokens = -1

	// First, extract prompt usage details from the request.
	promptUsage.characters = len(reqBody.Prompt)
	promptUsage.tokens, err = reqBody.GetPromptTokenCount(a.anthropicTokenizer)
	if err != nil {
		logger.Error("failed to count tokens in Anthropic response", log.Error(err))
	}
	promptUsage.tokenizerTokens = promptUsage.tokens

	// Try to parse the request we saw, if it was non-streaming, we can simply parse
	// it as JSON.
	if !isStreamRequest {
		var res anthropicResponse
		if err := json.NewDecoder(r).Decode(&res); err != nil {
			logger.Error("failed to parse Anthropic response as JSON", log.Error(err))
			return promptUsage, completionUsage
		}

		// Extract usage data from response
		completionUsage.characters = len(res.Completion)
		if tokens, err := a.anthropicTokenizer.Tokenize(res.Completion); err != nil {
			logger.Error("failed to count tokens in Anthropic response", log.Error(err))
		} else {
			completionUsage.tokens = len(tokens)
			completionUsage.tokenizerTokens = completionUsage.tokens
		}
		return promptUsage, completionUsage
	}

	// Otherwise, we have to parse the event stream from anthropic.
	dec := anthropic.NewDecoder(r)
	var lastCompletion string
	// Consume all the messages, but we only care about the last completion data.
	for dec.Scan() {
		data := dec.Data()

		// Gracefully skip over any data that isn't JSON-like. Anthropic's API sometimes sends
		// non-documented data over the stream, like timestamps.
		if !bytes.HasPrefix(data, []byte("{")) {
			continue
		}

		var event anthropicResponse
		if err := json.Unmarshal(data, &event); err != nil {
			logger.Error("failed to decode event payload", log.Error(err), log.String("body", string(data)))
			continue
		}
		lastCompletion = event.Completion
	}
	if err := dec.Err(); err != nil {
		logger.Error("failed to decode Anthropic streaming response", log.Error(err))
	}

	// Extract usage data from streamed response.
	completionUsage.characters = len(lastCompletion)
	if tokens, err := a.anthropicTokenizer.Tokenize(lastCompletion); err != nil {
		logger.Warn("failed to count tokens in Anthropic response", log.Error(err))
		completionUsage.tokens = -1
	} else {
		completionUsage.tokens = len(tokens)
	}
	completionUsage.tokenizerTokens = completionUsage.tokens
	return promptUsage, completionUsage
}
