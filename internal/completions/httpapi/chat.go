package httpapi

import (
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewChatCompletionsStreamHandler is an http handler which streams back completions results.
func NewChatCompletionsStreamHandler(logger log.Logger, db database.DB) http.Handler {
	logger = logger.Scoped("chat", "chat completions handler")
	rl := NewRateLimiter(db, redispool.Store, types.CompletionsFeatureChat)
	getConfig := func() conftypes.ProviderConfig {
		return conf.GetCompletionsConfig(conf.Get().SiteConfig())
	}

	return newCompletionsHandler(
		logger,
		types.CompletionsFeatureChat,
		rl,
		"chat",
		getConfig,
		func(requestParams types.CodyCompletionRequestParameters) (string, error) {
			config := conf.GetCompletionsConfig(conf.Get().SiteConfig())
			if config == nil {
				return "", errors.New("completions are not configured or disabled")
			}
			// No user defined models for now.
			if requestParams.Fast {
				return config.FastChatModel, nil
			}
			return config.ChatModel, nil
		},
	)
}
