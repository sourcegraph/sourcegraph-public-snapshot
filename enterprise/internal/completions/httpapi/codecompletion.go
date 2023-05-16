package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/schema"
)

var allowedClientSpecifiedModels = map[string]struct{}{
	// TODO(eseliger): This list should probably be configurable.
	"claude-instant-v1.0": {},
}

// NewCodeCompletionsHandler is an http handler which sends back code completion results
func NewCodeCompletionsHandler(_ log.Logger, db database.DB) http.Handler {
	rl := NewRateLimiter(db, redispool.Store, RateLimitScopeCodeCompletion)
	return newCompletionsHandler(rl, "codeCompletions", func(requestParams types.CompletionRequestParameters, c *schema.Completions) string {
		var model string
		if _, isAllowed := allowedClientSpecifiedModels[requestParams.Model]; isAllowed {
			model = requestParams.Model
		} else {
			model = c.CompletionModel
		}

		return model
	}, func(ctx context.Context, requestParams types.CompletionRequestParameters, cc types.CompletionsClient, w http.ResponseWriter) {
		completion, err := cc.Complete(ctx, requestParams)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		completionBytes, err := json.Marshal(completion)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(completionBytes)
	})
}
