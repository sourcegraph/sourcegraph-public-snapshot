package httpapi

import (
	"context"
	sgactor "github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
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
		func(ctx context.Context, requestParams types.CodyCompletionRequestParameters, c *conftypes.CompletionsConfig) (string, error) {
			actor := sgactor.FromContext(ctx)
			user, err := actor.User(ctx, db.Users())
			if err != nil {
				return "", err
			}
			isCodyProEnabled := featureflag.FromContext(ctx).GetBoolOr("cody-pro", false)
			isProUser := user.CodyProEnabledAt != nil
			if isAllowedCustomChatModel(requestParams.Model, isProUser || !isCodyProEnabled) {
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
func isAllowedCustomChatModel(model string, isProUser bool) bool {
	if !(envvar.SourcegraphDotComMode()) {
		return false
	}

	if isProUser {
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
	} else {
		if model == "anthropic/claude-2.0" {
			return true
		}
	}

	return false
}
