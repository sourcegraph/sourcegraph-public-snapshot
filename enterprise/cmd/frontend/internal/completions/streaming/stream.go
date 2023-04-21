// This package defines both streaming and non-streaming completion REST endpoints. Should probably be renamed "rest".
package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/completions/streaming/anthropic"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/completions/streaming/openai"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/cody"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const maxRequestDuration = time.Minute

// NewCompletionsStreamHandler is an http handler which streams back completions results.
func NewCompletionsStreamHandler(logger log.Logger, db database.DB) http.Handler {
	rl := NewRateLimiter(db, redispool.Store)
	return &streamHandler{logger: logger, rl: rl}
}

type streamHandler struct {
	logger log.Logger
	rl     RateLimiter
}

func GetCompletionClient(provider string, accessToken string, model string) (types.CompletionsClient, error) {
	switch provider {
	case "anthropic":
		return anthropic.NewAnthropicClient(httpcli.ExternalDoer, accessToken, model), nil
	case "openai":
		return openai.NewOpenAIClient(httpcli.ExternalDoer, accessToken, model), nil
	default:
		return nil, errors.Newf("unknown completion stream provider: %s", provider)
	}
}

func (h *streamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), maxRequestDuration)
	defer cancel()

	completionsConfig := conf.Get().Completions
	if completionsConfig == nil || !completionsConfig.Enabled {
		http.Error(w, "completions are not configured or disabled", http.StatusInternalServerError)
		return
	}

	if isEnabled := cody.IsCodyEnabled(ctx); !isEnabled {
		http.Error(w, "cody experimental feature flag is not enabled for current user", http.StatusUnauthorized)
		return
	}

	if r.Method != "POST" {
		http.Error(w, fmt.Sprintf("unsupported method %s", r.Method), http.StatusBadRequest)
		return
	}

	var requestParams types.ChatCompletionRequestParameters
	if err := json.NewDecoder(r.Body).Decode(&requestParams); err != nil {
		http.Error(w, "could not decode request body", http.StatusBadRequest)
		return
	}

	model := completionsConfig.ChatModel
	if model == "" {
		model = completionsConfig.Model
	}

	var err error
	ctx, done := Trace(ctx, "stream", model).
		WithErrorP(&err).
		WithRequest(r).
		Build()
	defer done()

	completionClient, err := GetCompletionClient(completionsConfig.Provider, completionsConfig.AccessToken, model)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check rate limit.
	err = h.rl.TryAcquire(ctx)
	if err != nil {
		if unwrap, ok := err.(RateLimitExceededError); ok {
			respondRateLimited(w, unwrap)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	eventWriter, err := streamhttp.NewWriter(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Always send a final done event so clients know the stream is shutting down.
	defer eventWriter.Event("done", map[string]any{})

	err = completionClient.Stream(ctx, requestParams, func(event types.ChatCompletionEvent) error { return eventWriter.Event("completion", event) })
	if err != nil {
		h.logger.Error("error while streaming completions", log.Error(err))
		eventWriter.Event("error", map[string]string{"error": err.Error()})
		return
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func respondRateLimited(w http.ResponseWriter, err RateLimitExceededError) {
	// Rate limit exceeded, write well known headers and return correct status code.
	w.Header().Set("x-ratelimit-limit", strconv.Itoa(err.Limit))
	w.Header().Set("x-ratelimit-remaining", strconv.Itoa(max(err.Limit-err.Used, 0)))
	w.Header().Set("retry-after", err.RetryAfter.Format(time.RFC3339))
	http.Error(w, err.Error(), http.StatusTooManyRequests)
}
