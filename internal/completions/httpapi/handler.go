package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/cody"
	"github.com/sourcegraph/sourcegraph/internal/completions/client"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

// maxRequestDuration is the maximum amount of time a request can take before
// being cancelled.
const maxRequestDuration = time.Minute

func newCompletionsHandler(
	rl RateLimiter,
	traceFamily string,
	getModel func(types.CodyCompletionRequestParameters, *conftypes.CompletionsConfig) string,
	handle func(context.Context, types.CompletionRequestParameters, types.CompletionsClient, http.ResponseWriter),
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, fmt.Sprintf("unsupported method %s", r.Method), http.StatusMethodNotAllowed)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), maxRequestDuration)
		defer cancel()

		if isEnabled := cody.IsCodyEnabled(ctx); !isEnabled {
			http.Error(w, "cody experimental feature flag is not enabled for current user", http.StatusUnauthorized)
			return
		}

		completionsConfig := conf.GetCompletionsConfig(conf.Get().SiteConfig())
		if completionsConfig == nil {
			http.Error(w, "completions are not configured or disabled", http.StatusInternalServerError)
		}

		var requestParams types.CodyCompletionRequestParameters
		if err := json.NewDecoder(r.Body).Decode(&requestParams); err != nil {
			http.Error(w, "could not decode request body", http.StatusBadRequest)
			return
		}

		// TODO: Model is not configurable but technically allowed in the request body right now.
		requestParams.Model = getModel(requestParams, completionsConfig)

		var err error
		ctx, done := Trace(ctx, traceFamily, requestParams.Model, requestParams.MaxTokensToSample).
			WithErrorP(&err).
			WithRequest(r).
			Build()
		defer done()

		completionClient, err := client.Get(
			completionsConfig.Endpoint,
			completionsConfig.Provider,
			completionsConfig.AccessToken,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Check rate limit.
		err = rl.TryAcquire(ctx)
		if err != nil {
			if unwrap, ok := err.(RateLimitExceededError); ok {
				respondRateLimited(w, unwrap)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		handle(ctx, requestParams.CompletionRequestParameters, completionClient, w)
	})
}

func respondRateLimited(w http.ResponseWriter, err RateLimitExceededError) {
	// Rate limit exceeded, write well known headers and return correct status code.
	w.Header().Set("x-ratelimit-limit", strconv.Itoa(err.Limit))
	w.Header().Set("x-ratelimit-remaining", strconv.Itoa(max(err.Limit-err.Used, 0)))
	w.Header().Set("retry-after", err.RetryAfter.Format(time.RFC1123))
	http.Error(w, err.Error(), http.StatusTooManyRequests)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
