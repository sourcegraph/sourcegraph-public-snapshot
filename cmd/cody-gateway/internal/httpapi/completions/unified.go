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

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/notify"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

const anthropicMessagesAPIURL = "https://api.anthropic.com/v1/messages"

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
	return makeUpstreamHandler[unifiedRequest](
		baseLogger,
		eventLogger,
		rs,
		rateLimitNotifier,
		httpClient,
		string(conftypes.CompletionsProviderNameAnthropic),
		func(_ codygateway.Feature) string { return anthropicMessagesAPIURL },
		config.AllowedModels,
		&UnifiedHandlerMethods{config: config},

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
	Messages      []unifiedMessage        `json:"messages,omitempty"`
	Model         string                  `json:"model"`
	MaxTokens     int32                   `json:"max_tokens,omitempty"`
	Temperature   float32                 `json:"temperature,omitempty"`
	TopP          float32                 `json:"top_p,omitempty"`
	TopK          int32                   `json:"top_k,omitempty"`
	Stream        bool                    `json:"stream,omitempty"`
	StopSequences []string                `json:"stop_sequences,omitempty"`
	Metadata      *unifiedRequestMetadata `json:"metadata,omitempty"`
}

type unifiedMessage struct {
	Role    string           `json:"role"`
	Content []unifiedContent `json:"content"`
}

type unifiedContent struct {
	Type string `json:"type"` // "text" or "image_url"

	Text string `json:"text",omitempty"`
	URL  string `json:"url,omitempty"`
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

func (r unifiedRequest) BuildPrompt() string {
	var sb strings.Builder
	for _, m := range r.Messages {
		for _, c := range m.Content {
			if c.Type == "text" {
				sb.WriteString(c.Text + "\n")
			}
		}
	}
	return sb.String()
}

// unifiedResponse captures all relevant-to-us fields from https://console.anthropic.com/docs/api/reference.
type unifiedResponse struct {
	Completion string `json:"completion,omitempty"`
	StopReason string `json:"stop_reason,omitempty"`
}

type UnifiedHandlerMethods struct {
	config config.AnthropicConfig
}

func (a *UnifiedHandlerMethods) validateRequest(ctx context.Context, logger log.Logger, _ codygateway.Feature, ar unifiedRequest) (int, *flaggingResult, error) {
	return 0, nil, nil
}
func (a *UnifiedHandlerMethods) transformBody(body *unifiedRequest, identifier string) {
	// Overwrite the metadata field, we don't want to allow users to specify it:
	body.Metadata = &unifiedRequestMetadata{
		// We forward the actor ID to support tracking.
		UserID: identifier,
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

	// Try to parse the request we saw, if it was non-streaming, we can simply parse
	// it as JSON.
	if !body.Stream {
		var res unifiedResponse
		if err := json.NewDecoder(r).Decode(&res); err != nil {
			logger.Error("failed to parse Anthropic response as JSON", log.Error(err))
			return promptUsage, completionUsage
		}

		// Extract usage data from response
		completionUsage.characters = len(res.Completion)

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

		var event unifiedResponse
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
	return promptUsage, completionUsage
}
