package httpapi

import (
	"context"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/telemetryrecorder"
)

// NewChatCompletionsStreamHandler is an http handler which streams back completions results.
func NewChatCompletionsStreamHandler(logger log.Logger, db database.DB) http.Handler {
	logger = logger.Scoped("chat")
	rl := NewRateLimiter(db, redispool.Store, types.CompletionsFeatureChat)

	return newCompletionsHandler(
		logger,
		db.Users(),
		db.AccessTokens(),
		telemetryrecorder.New(db),
		types.CompletionsFeatureChat,
		rl,
		"chat",
		func(_ context.Context, requestParams types.CodyCompletionRequestParameters, c *conftypes.CompletionsConfig) (string, error) {
			if isAllowedCustomChatModel(requestParams.Model) {
				return requestParams.Model, nil
			}
			// No user defined models for now.
			if requestParams.Fast {
				return c.FastChatModel, nil
			}
			return c.ChatModel, nil
		},
	)
}

// We only allow dotcom clients to select a custom chat model and maintain an allowlist for which
// custom values we support
func isAllowedCustomChatModel(model string) bool {
	if !(envvar.SourcegraphDotComMode()) {
		return false
	}

	switch model {
	case "anthropic/claude-2",
		"anthropic/claude-2.0",
		"anthropic/claude-2.1",
		"anthropic/claude-instant-1.2-cyan",
		"anthropic/claude-instant-1.2",
		"openai/gpt-3.5-turbo",
		"openai/gpt-4-1106-preview":
		return true
	}

	return false
}
