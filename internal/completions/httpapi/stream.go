package httpapi

import (
	"context"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// NewChatCompletionsStreamHandler is an http handler which streams back completions results.
func NewChatCompletionsStreamHandler(logger log.Logger, db database.DB) http.Handler {
	logger = logger.Scoped("chat", "chat completions handler")
	rl := NewRateLimiter(db, redispool.Store, types.CompletionsFeatureChat)

	return newCompletionsHandler(rl, "chat", func(requestParams types.CodyCompletionRequestParameters, c *conftypes.CompletionsConfig) string {
		// No user defined models for now.
		if requestParams.Fast {
			return c.FastChatModel
		}
		return c.ChatModel
	}, func(ctx context.Context, requestParams types.CompletionRequestParameters, cc types.CompletionsClient, w http.ResponseWriter) {
		eventWriter, err := streamhttp.NewWriter(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Always send a final done event so clients know the stream is shutting down.
		defer func() {
			_ = eventWriter.Event("done", map[string]any{})
		}()

		err = cc.Stream(ctx, types.CompletionsFeatureChat, requestParams,
			func(event types.CompletionResponse) error {
				return eventWriter.Event("completion", event)
			})
		if err != nil {
			l := trace.Logger(ctx, logger)

			logFields := []log.Field{log.Error(err)}
			if errNotOK, ok := types.IsErrStatusNotOK(err); ok {
				if tc := errNotOK.SourceTraceContext; tc != nil {
					logFields = append(logFields,
						log.String("sourceTraceContext.traceID", tc.TraceID),
						log.String("sourceTraceContext.spanID", tc.SpanID))
				}
			}
			l.Error("error while streaming completions", logFields...)

			// Note that we do NOT attempt to forward the status code to the
			// client here, since we are using streamhttp.Writer - see
			// streamhttp.NewWriter for more details. Instead, we send an error
			// event, which clients should check as appropriate.
			if err := eventWriter.Event("error", map[string]string{"error": err.Error()}); err != nil {
				l.Error("error reporting streaming completion error", log.Error(err))
			}
			return
		}
	})
}
