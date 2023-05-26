package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/cody"
	"github.com/sourcegraph/sourcegraph/schema"
)

// maxRequestDuration is the maximum amount of time a request can take before
// being cancelled.
const maxRequestDuration = time.Minute

func newCompletionsHandler(rl RateLimiter, traceFamily string, getModel func(types.CompletionRequestParameters, *schema.Completions) string, handle func(context.Context, types.CompletionRequestParameters, types.CompletionsClient, http.ResponseWriter)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, fmt.Sprintf("unsupported method %s", r.Method), http.StatusMethodNotAllowed)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), maxRequestDuration)
		defer cancel()

		completionsConfig := client.GetCompletionsConfig()
		if completionsConfig == nil || !completionsConfig.Enabled {
			http.Error(w, "completions are not configured or disabled", http.StatusInternalServerError)
			return
		}

		if isEnabled := cody.IsCodyEnabled(ctx); !isEnabled {
			http.Error(w, "cody experimental feature flag is not enabled for current user", http.StatusUnauthorized)
			return
		}

		var requestParams types.CompletionRequestParameters
		if err := json.NewDecoder(r.Body).Decode(&requestParams); err != nil {
			http.Error(w, "could not decode request body", http.StatusBadRequest)
			return
		}

		// TODO: Model is not configurable but technically allowed in the request body right now.
		requestParams.Model = getModel(requestParams, completionsConfig)

		var err error
		ctx, done := Trace(ctx, traceFamily, requestParams.Model).
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
			if unwrap, ok := err.(types.RateLimitExceededError); ok {
				unwrap.WriteHTTPResponse(w)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		handle(ctx, requestParams, completionClient, w)
	})
}
