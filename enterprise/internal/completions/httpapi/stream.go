package httpapi

import (
	"context"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewCompletionsStreamHandler is an http handler which streams back completions results.
func NewCompletionsStreamHandler(logger log.Logger, db database.DB) http.Handler {
	rl := NewRateLimiter(db, redispool.Store, RateLimitScopeCompletion)
	return newCompletionsHandler(rl, "stream", func(c *schema.Completions) string {
		return c.ChatModel
	}, func(ctx context.Context, requestParams types.ChatCompletionRequestParameters, cc types.CompletionsClient, w http.ResponseWriter) {
		eventWriter, err := streamhttp.NewWriter(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Always send a final done event so clients know the stream is shutting down.
		defer func() {
			_ = eventWriter.Event("done", map[string]any{})
		}()

		err = cc.Stream(ctx, requestParams, func(event types.ChatCompletionEvent) error { return eventWriter.Event("completion", event) })
		if err != nil {
			trace.Logger(ctx, logger).Error("error while streaming completions", log.Error(err))
			_ = eventWriter.Event("error", map[string]string{"error": err.Error()})
			return
		}
	})
}
