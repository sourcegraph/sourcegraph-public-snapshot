package completions

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/shared/config"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/anthropic"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/notify"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/tokenizer"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

const anthropicMessagesAPIURL = "https://api.anthropic.com/v1/messages"

// The unified API endpoint is a new general-purpose AI inference API inspired
// by the OpenAI API. The idea is is that regardless of which model you want to
// use, there's one _unified_ API to use for all Sourcegraph products.
//
// Sharing the same API interface between the SG instance and Cody Gateway also
// allows clients to implement RFC888 and connect to Cody Gateway directly.
//
// Right now, unified API is only available for the Claude 3 model family and
// only implemented on the Cody Gateway side for PLG users connecting directly.
func NewUnifiedHandler(
	baseLogger log.Logger,
	eventLogger events.Logger,
	rs limiter.RedisStore,
	rateLimitNotifier notify.RateLimitNotifier,
	httpClient httpcli.Doer,
	config config.AnthropicConfig,
	promptRecorder PromptRecorder,
	autoFlushStreamingResponses bool,
) (http.Handler, error) {
	// Tokenizer only needs to be initialized once, and can be shared globally.
	tokenizer, err := tokenizer.NewAnthropicClaudeTokenizer()
	if err != nil {
		return nil, err
	}
	return makeUpstreamHandler[unifiedRequest](
		baseLogger,
		eventLogger,
		rs,
		rateLimitNotifier,
		httpClient,
		string(conftypes.CompletionsProviderNameAnthropic),
		func(_ codygateway.Feature) string { return anthropicMessagesAPIURL },
		config.AllowedModels,
		&UnifiedHandlerMethods{config: config, tokenizer: tokenizer, promptRecorder: promptRecorder},

		// Anthropic primarily uses concurrent requests to rate-limit spikes
		// in requests, so set a default retry-after that is likely to be
		// acceptable for Sourcegraph clients to retry (the default
		// SRC_HTTP_CLI_EXTERNAL_RETRY_AFTER_MAX_DURATION) since we might be
		// able to circumvent concurrents limits without raising an error to the
		// user.
		2, // seconds
		autoFlushStreamingResponses,
		nil,
	), nil
}

// unifiedRequest captures all known fields from https://console.anthropic.com/docs/api/reference.
type unifiedRequest struct {
	Messages      []unifiedMessage `json:"messages,omitempty"`
	Model         string           `json:"model"`
	MaxTokens     int32            `json:"max_tokens,omitempty"`
	Temperature   float32          `json:"temperature,omitempty"`
	TopP          float32          `json:"top_p,omitempty"`
	TopK          int32            `json:"top_k,omitempty"`
	Stream        bool             `json:"stream,omitempty"`
	StopSequences []string         `json:"stop_sequences,omitempty"`

	// TODO: These are not accepted from the client an instead are only used to
	// talk to the upstream LLM APIs.
	Metadata *unifiedRequestMetadata `json:"metadata,omitempty"`
	System   string                  `json:"system,omitempty"`
}

type unifiedMessage struct {
	Role    string           `json:"role"` // "user", "assistant", or "system" (only allowed for the first message)
	Content []unifiedContent `json:"content"`
}

type unifiedContent struct {
	Type string `json:"type"` // "text" or "image" (not yet supported)
	Text string `json:"text"`
}

type unifiedRequestMetadata struct {
	UserID string `json:"user_id,omitempty"`
}

func (ar unifiedRequest) ShouldStream() bool {
	return ar.Stream
}

func (ar unifiedRequest) GetModel() string {
	return ar.Model
}

// Note: This is not the actual prompt send to Anthropic but it's a good
// approximation to measure tokens.
func (r unifiedRequest) BuildPrompt() string {
	var sb strings.Builder
	for _, m := range r.Messages {
		switch m.Role {
		case "user":
			sb.WriteString("Human: ")
		case "assistant":
			sb.WriteString("Assistant: ")
		case "system":
			sb.WriteString("System: ")
		default:
			return ""
		}

		for _, c := range m.Content {
			if c.Type == "text" {
				sb.WriteString(c.Text)
			}
		}
		sb.WriteString("\n\n")
	}
	return sb.String()
}

// GetPromptTokenCount computes the token count of the prompt exactly once using
// the given tokenizer. It is not concurrency-safe.
func (r *unifiedRequest) GetPromptTokenCount(tk *tokenizer.Tokenizer) (int, error) {
	tokens, err := tk.Tokenize(r.BuildPrompt())
	return len(tokens), err
}

// unifiedNonStreamingResponse captures all relevant-to-us fields from https://docs.anthropic.com/claude/reference/messages_post.
type unifiedNonStreamingResponse struct {
	Content    []unifiedContent     `json:"content"`
	Usage      unifiedResponseUsage `json:"usage"`
	StopReason string               `json:"stop_reason"`
}

// unifiedStreamingResponse captures all relevant-to-us fields from each relevant SSE event from https://docs.anthropic.com/claude/reference/messages_post.
type unifiedStreamingResponse struct {
	Type         string                              `json:"type"`
	Delta        *unifiedStreamingResponseTextBucket `json:"delta"`
	ContentBlock *unifiedStreamingResponseTextBucket `json:"content_block"`
	Usage        *unifiedResponseUsage               `json:"usage"`
}

type unifiedStreamingResponseTextBucket struct {
	Text string `json:"text"`
}

type unifiedResponseUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type UnifiedHandlerMethods struct {
	tokenizer      *tokenizer.Tokenizer
	promptRecorder PromptRecorder
	config         config.AnthropicConfig
}

func (a *UnifiedHandlerMethods) validateRequest(ctx context.Context, logger log.Logger, _ codygateway.Feature, ar unifiedRequest) (int, *flaggingResult, error) {
	if ar.MaxTokens > int32(a.config.MaxTokensToSample) {
		return http.StatusBadRequest, nil, errors.Errorf("max_tokens exceeds maximum allowed value of %d: %d", a.config.MaxTokensToSample, ar.MaxTokens)
	}

	if result, err := isFlaggedUnifiedRequest(a.tokenizer, ar, a.config); err != nil {
		logger.Error("error checking unified request - treating as non-flagged",
			log.Error(err))
	} else if result.IsFlagged() {
		// Record flagged prompts in hotpath - they usually take a long time on the backend side, so this isn't going to make things meaningfully worse
		if err := a.promptRecorder.Record(ctx, ar.BuildPrompt()); err != nil {
			logger.Warn("failed to record flagged prompt", log.Error(err))
		}
		if a.config.RequestBlockingEnabled && result.shouldBlock {
			return http.StatusBadRequest, result, errors.Errorf("request blocked - if you think this is a mistake, please contact support@sourcegraph.com")
		}
		return 0, result, nil
	}

	return 0, nil, nil
}
func (a *UnifiedHandlerMethods) transformBody(body *unifiedRequest, identifier string) {
	// Overwrite the metadata field, we don't want to allow users to specify it:
	body.Metadata = &unifiedRequestMetadata{
		// We forward the actor ID to support tracking.
		UserID: identifier,
	}

	// Remove the `anthropic/` prefix from the model string
	body.Model = strings.TrimPrefix(body.Model, "anthropic/")

	// Convert the eventual first message from `system` to a top-level system prompt
	body.System = "" // prevent the upstream API from setting this
	if len(body.Messages) > 0 && body.Messages[0].Role == "system" {
		body.System = body.Messages[0].Content[0].Text
		body.Messages = body.Messages[1:]
	}
}
func (a *UnifiedHandlerMethods) getRequestMetadata(body unifiedRequest) (model string, additionalMetadata map[string]any) {
	return body.Model, map[string]any{
		"stream":     body.Stream,
		"max_tokens": body.MaxTokens,
	}
}
func (a *UnifiedHandlerMethods) transformRequest(r *http.Request) {
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-API-Key", a.config.AccessToken)
	r.Header.Set("anthropic-version", "2023-06-01")
}
func (a *UnifiedHandlerMethods) parseResponseAndUsage(logger log.Logger, body unifiedRequest, r io.Reader) (promptUsage, completionUsage usageStats) {
	// First, extract prompt usage details from the request.
	for _, m := range body.Messages {
		promptUsage.characters += len(m.Content)
	}
	promptUsage.characters = len(body.System)

	// Try to parse the request we saw, if it was non-streaming, we can simply parse
	// it as JSON.
	if !body.ShouldStream() {
		var res unifiedNonStreamingResponse
		if err := json.NewDecoder(r).Decode(&res); err != nil {
			logger.Error("failed to parse Anthropic response as JSON", log.Error(err))
			return promptUsage, completionUsage
		}

		// Extract character data from response by summing up all text
		for _, c := range res.Content {
			completionUsage.characters += len(c.Text)
		}
		// Extract prompt usage data from the response
		completionUsage.tokens = res.Usage.OutputTokens
		promptUsage.tokens = res.Usage.InputTokens

		return promptUsage, completionUsage
	}

	// Otherwise, we have to parse the event stream from anthropic.
	dec := anthropic.NewDecoder(r)
	for dec.Scan() {
		data := dec.Data()

		// Gracefully skip over any data that isn't JSON-like. Anthropic's API sometimes sends
		// non-documented data over the stream, like timestamps.
		if !bytes.HasPrefix(data, []byte("{")) {
			continue
		}

		var event unifiedStreamingResponse
		if err := json.Unmarshal(data, &event); err != nil {
			logger.Error("failed to decode event payload", log.Error(err), log.String("body", string(data)))
			continue
		}

		switch event.Type {
		case "message_start":
			if event.Usage != nil {
				promptUsage.tokens = event.Usage.InputTokens
			}
		case "content_block_delta":
			if event.Delta != nil {
				completionUsage.characters += len(event.Delta.Text)
			}
		case "message_delta":
			if event.Usage != nil {
				completionUsage.tokens = event.Usage.OutputTokens
			}
		}
	}
	if err := dec.Err(); err != nil {
		logger.Error("failed to decode Anthropic streaming response", log.Error(err))
	}

	return promptUsage, completionUsage
}

func isFlaggedUnifiedRequest(tk *tokenizer.Tokenizer, r unifiedRequest, cfg config.AnthropicConfig) (*flaggingResult, error) {
	var reasons []string

	if len(cfg.AllowedPromptPatterns) > 0 && !containsAny(r.BuildPrompt(), cfg.AllowedPromptPatterns) {
		reasons = append(reasons, "unknown_prompt")
	}

	// If this request has a very high token count for responses, then flag it.
	if r.MaxTokens > int32(cfg.MaxTokensToSampleFlaggingLimit) {
		reasons = append(reasons, "high_max_tokens_to_sample")
	}

	// If this prompt consists of a very large number of tokens, then flag it.
	tokenCount, err := r.GetPromptTokenCount(tk)
	if err != nil {
		return &flaggingResult{}, errors.Wrap(err, "tokenize prompt")
	}
	if tokenCount > cfg.PromptTokenFlaggingLimit {
		reasons = append(reasons, "high_prompt_token_count")
	}

	if len(reasons) > 0 {
		blocked := false
		if tokenCount > cfg.PromptTokenBlockingLimit || r.MaxTokens > int32(cfg.ResponseTokenBlockingLimit) || containsAny(r.BuildPrompt(), cfg.BlockedPromptPatterns) {
			blocked = true
		}

		promptPrefix := r.BuildPrompt()
		if len(promptPrefix) > logPromptPrefixLength {
			promptPrefix = promptPrefix[0:logPromptPrefixLength]
		}
		return &flaggingResult{
			reasons:           reasons,
			maxTokensToSample: int(r.MaxTokens),
			promptPrefix:      promptPrefix,
			promptTokenCount:  tokenCount,
			shouldBlock:       blocked,
		}, nil
	}

	return nil, nil
}
