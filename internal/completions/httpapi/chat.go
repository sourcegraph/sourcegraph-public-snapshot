package httpapi

import (
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/telemetryrecorder"
)

// NewChatCompletionsStreamHandler is an http handler which streams back completions results.
func NewChatCompletionsStreamHandler(logger log.Logger, db database.DB) http.Handler {
	logger = logger.Scoped("chat", "chat completions handler")
	rl := NewRateLimiter(db, redispool.Store, types.CompletionsFeatureChat)

	return newCompletionsHandler(
		logger,
		telemetryrecorder.New(db),
		types.CompletionsFeatureChat,
		rl,
		"chat",
		func(requestParams types.CodyCompletionRequestParameters, c *conftypes.CompletionsConfig) (string, error) {
			// No user defined models for now.
			if requestParams.Fast {
				return c.FastChatModel, nil
			}
			return c.ChatModel, nil
		},
	)
}
