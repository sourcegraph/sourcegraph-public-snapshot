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

// NewCodeCompletionsHandler is an http handler which sends back code completion results
func NewCodeCompletionsHandler(_ log.Logger, db database.DB) http.Handler {
	rl := NewRateLimiter(db, redispool.Store, RateLimitScopeCodeCompletion)
	return newCompletionsHandler(rl, "codeCompletions", func(c *schema.Completions) string {
		return c.CompletionModel
	}, func(ctx context.Context, requestParams types.CodeCompletionRequestParameters, cc types.CompletionsClient, w http.ResponseWriter) {
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
