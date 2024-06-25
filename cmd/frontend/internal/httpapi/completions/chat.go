package completions

import (
	"context"

	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/telemetryrecorder"
)

// chatAttributionTest always returns true, as chat attribution
// is performed on the client side (as opposed to code completions)
// which works on the server side.
func chatAttributionTest(context.Context, string) (bool, error) {
	return true, nil
}

// NewChatCompletionsStreamHandler is an http handler which streams back completions results.
func NewChatCompletionsStreamHandler(logger log.Logger, db database.DB) http.Handler {
	logger = logger.Scoped("chat")
	rl := NewRateLimiter(db, redispool.Store, types.CompletionsFeatureChat)

	return newCompletionsHandler(
		logger,
		db,
		db.Users(),
		db.AccessTokens(),
		telemetryrecorder.New(db),
		chatAttributionTest,
		types.CompletionsFeatureChat,
		rl,
		"chat",
		getChatModelFn(db))
}
