package completions

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/actor"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/notify"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/tokenizer"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/anthropic"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const anthropicAPIURL = "https://api.anthropic.com/v1/complete"

const (
	logPromptPrefixLength = 250

	promptTokenFlaggingLimit   = 18000
	responseTokenFlaggingLimit = 1000

	promptTokenBlockingLimit   = 20000
	responseTokenBlockingLimit = 1000
)

func isFlaggedAnthropicRequest(tk *tokenizer.Tokenizer, ar anthropicRequest, promptRegexps []*regexp.Regexp) (*flaggingResult, error) {
	// Only usage of chat models us currently flagged, so if the request
	// is using another model, we skip other checks.
	if ar.Model != "claude-2" && ar.Model != "claude-2.0" && ar.Model != "claude-2.1" && ar.Model != "claude-v1" {
		return nil, nil
	}
	reasons := []string{}

	if len(promptRegexps) > 0 && !matchesAny(ar.Prompt, promptRegexps) {
		reasons = append(reasons, "unknown_prompt")
	}

	// If this request has a very high token count for responses, then flag it.
	if ar.MaxTokensToSample > responseTokenFlaggingLimit {
		reasons = append(reasons, "high_max_tokens_to_sample")
	}

	// If this prompt consists of a very large number of tokens, then flag it.
	tokenCount, err := ar.GetPromptTokenCount(tk)
	if err != nil {
		return &flaggingResult{}, errors.Wrap(err, "tokenize prompt")
	}
	if tokenCount > promptTokenFlaggingLimit {
		reasons = append(reasons, "high_prompt_token_count")
	}

	if len(reasons) > 0 {
		blocked := false
		if tokenCount > promptTokenBlockingLimit || ar.MaxTokensToSample > responseTokenBlockingLimit {
			blocked = true
		}

		promptPrefix := ar.Prompt
		if len(promptPrefix) > logPromptPrefixLength {
			promptPrefix = promptPrefix[0:logPromptPrefixLength]
		}
		return &flaggingResult{
			reasons:           reasons,
			maxTokensToSample: int(ar.MaxTokensToSample),
			promptPrefix:      promptPrefix,
			promptTokenCount:  tokenCount,
			shouldBlock:       blocked,
		}, nil
	}

	return nil, nil
}

func matchesAny(prompt string, promptRegexps []*regexp.Regexp) bool {
	for _, promptRegexp := range promptRegexps {
		if promptRegexp.MatchString(prompt) {
			return true
		}
	}
	return false
}

// PromptRecorder implementations should save select completions prompts for
// a short amount of time for security review.
type PromptRecorder interface {
	Record(ctx context.Context, prompt string) error
}

func NewAnthropicHandler(
	baseLogger log.Logger,
	eventLogger events.Logger,
	rs limiter.RedisStore,
	rateLimitNotifier notify.RateLimitNotifier,
	httpClient httpcli.Doer,
	accessToken string,
	allowedModels []string,
	maxTokensToSample int,
	promptRecorder PromptRecorder,
	allowedPromptPatterns []string,
	requestBlockingEnabled bool,
	autoFlushStreamingResponses bool,
) (http.Handler, error) {
	// Tokenizer only needs to be initialized once, and can be shared globally.
	anthropicTokenizer, err := tokenizer.NewAnthropicClaudeTokenizer()
	if err != nil {
		return nil, err
	}
	promptRegexps := []*regexp.Regexp{}
	for _, pattern := range allowedPromptPatterns {
		promptRegexps = append(promptRegexps, regexp.MustCompile(pattern))
	}
	return makeUpstreamHandler(
		baseLogger,
		eventLogger,
		rs,
		rateLimitNotifier,
		httpClient,
		string(conftypes.CompletionsProviderNameAnthropic),
		anthropicAPIURL,
		allowedModels,
		upstreamHandlerMethods[anthropicRequest]{
			validateRequest: func(ctx context.Context, logger log.Logger, _ codygateway.Feature, ar anthropicRequest) (int, *flaggingResult, error) {
				if ar.MaxTokensToSample > int32(maxTokensToSample) {
					return http.StatusBadRequest, nil, errors.Errorf("max_tokens_to_sample exceeds maximum allowed value of %d: %d", maxTokensToSample, ar.MaxTokensToSample)
				}

				if result, err := isFlaggedAnthropicRequest(anthropicTokenizer, ar, promptRegexps); err != nil {
					logger.Error("error checking anthropic request - treating as non-flagged",
						log.Error(err))
				} else if result.IsFlagged() {
					// Record flagged prompts in hotpath - they usually take a long time on the backend side, so this isn't going to make things meaningfully worse
					if err := promptRecorder.Record(ctx, ar.Prompt); err != nil {
						logger.Warn("failed to record flagged prompt", log.Error(err))
					}
					if requestBlockingEnabled && result.shouldBlock {
						return http.StatusBadRequest, result, errors.Errorf("request blocked - if you think this is a mistake, please contact support@sourcegraph.com")
					}
					return 0, result, nil
				}

				return 0, nil, nil
			},
			transformBody: func(body *anthropicRequest, identifier string) {
				// Overwrite the metadata field, we don't want to allow users to specify it:
				body.Metadata = &anthropicRequestMetadata{
					// We forward the actor ID to support tracking.
					UserID: identifier,
				}
			},
			getRequestMetadata: func(_ context.Context, _ log.Logger, _ *actor.Actor, body anthropicRequest) (model string, additionalMetadata map[string]any) {
				return body.Model, map[string]any{
					"stream":               body.Stream,
					"max_tokens_to_sample": body.MaxTokensToSample,
				}
			},
			transformRequest: func(r *http.Request) {
				// Mimic headers set by the official Anthropic client:
				// https://sourcegraph.com/github.com/anthropics/anthropic-sdk-typescript@493075d70f50f1568a276ed0cb177e297f5fef9f/-/blob/src/index.ts
				r.Header.Set("Cache-Control", "no-cache")
				r.Header.Set("Accept", "application/json")
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Client", "sourcegraph-cody-gateway/1.0")
				r.Header.Set("X-API-Key", accessToken)
				r.Header.Set("anthropic-version", "2023-01-01")
			},
			parseResponseAndUsage: func(logger log.Logger, reqBody anthropicRequest, r io.Reader) (promptUsage, completionUsage usageStats) {
				// First, extract prompt usage details from the request.
				promptUsage.characters = len(reqBody.Prompt)
				promptUsage.tokens, err = reqBody.GetPromptTokenCount(anthropicTokenizer)
				if err != nil {
					logger.Error("failed to count tokens in Anthropic response", log.Error(err))
				}

				// Try to parse the request we saw, if it was non-streaming, we can simply parse
				// it as JSON.
				if !reqBody.Stream {
					var res anthropicResponse
					if err := json.NewDecoder(r).Decode(&res); err != nil {
						logger.Error("failed to parse Anthropic response as JSON", log.Error(err))
						return promptUsage, completionUsage
					}

					// Extract usage data from response
					completionUsage.characters = len(res.Completion)
					if tokens, err := anthropicTokenizer.Tokenize(res.Completion); err != nil {
						logger.Error("failed to count tokens in Anthropic response", log.Error(err))
					} else {
						completionUsage.tokens = len(tokens)
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
						baseLogger.Error("failed to decode event payload", log.Error(err), log.String("body", string(data)))
						continue
					}
					lastCompletion = event.Completion
				}
				if err := dec.Err(); err != nil {
					logger.Error("failed to decode Anthropic streaming response", log.Error(err))
				}

				// Extract usage data from streamed response.
				completionUsage.characters = len(lastCompletion)
				if tokens, err := anthropicTokenizer.Tokenize(lastCompletion); err != nil {
					logger.Warn("failed to count tokens in Anthropic response", log.Error(err))
					completionUsage.tokens = -1
				} else {
					completionUsage.tokens = len(tokens)
				}
				return promptUsage, completionUsage
			},
		},

		// Anthropic primarily uses concurrent requests to rate-limit spikes
		// in requests, so set a default retry-after that is likely to be
		// acceptable for Sourcegraph clients to retry (the default
		// SRC_HTTP_CLI_EXTERNAL_RETRY_AFTER_MAX_DURATION) since we might be
		// able to circumvent concurrents limits without raising an error to the
		// user.
		2, // seconds
		autoFlushStreamingResponses,
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

type anthropicTokenCount struct {
	count int
	err   error
}

// GetPromptTokenCount computes the token count of the prompt exactly once using
// the given tokenizer. It is not concurrency-safe.
func (ar *anthropicRequest) GetPromptTokenCount(tk *tokenizer.Tokenizer) (int, error) {
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
