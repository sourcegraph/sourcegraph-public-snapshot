package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/cody"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/streaming/anthropic"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

// NewCodeCompletionsHandler is an http handler which sends back code completion results
func NewCodeCompletionsHandler(logger log.Logger, db database.DB) http.Handler {
	rl := NewRateLimiter(db, redispool.Store)
	return &codeCompletionHandler{logger: logger, rl: rl}
}

type codeCompletionHandler struct {
	logger log.Logger
	rl     RateLimiter
}

func (h *codeCompletionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), MaxRequestDuration)
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

	var p types.CodeCompletionRequestParameters
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var err error
	ctx, done := Trace(ctx, "codeCompletions", completionsConfig.CompletionModel).
		WithErrorP(&err).
		WithRequest(r).
		Build()
	defer done()

	client := anthropic.NewAnthropicClient(httpcli.ExternalDoer, completionsConfig.AccessToken, completionsConfig.CompletionModel)

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

	completion, err := client.Complete(ctx, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	completionBytes, err := json.Marshal(completion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(completionBytes)
}
