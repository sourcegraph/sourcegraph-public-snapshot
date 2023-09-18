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

var chatConfig *conftypes.CompletionsConfig

// NewChatCompletionsStreamHandler is an http handler which streams back completions results.
func NewChatCompletionsStreamHandler(logger log.Logger, db database.DB) http.Handler {
	logger = logger.Scoped("chat", "chat completions handler")
	rl := NewRateLimiter(db, redispool.Store, types.CompletionsFeatureChat)
	chatConfig = conf.GetChatCompletionsConfig(conf.Get().SiteConfig())
	go conf.Watch(func() {
		oldProvider := chatConfig.ProviderName()
		chatConfig = conf.GetChatCompletionsConfig(conf.Get().SiteConfig())
		logger.Info("Updating chat config", log.String("Old Provider", string(oldProvider)), log.String("New Provider", string(chatConfig.ProviderName())))
	})

	return newCompletionsHandler(
		logger,
		types.CompletionsFeatureChat,
		rl,
		"chat",
		chatConfig,
		func(requestParams types.CodyCompletionRequestParameters) (string, error) {
			if chatConfig == nil {
				return "", errors.New("completions are not configured or disabled")
			}
			// No user defined models for now.
			if requestParams.Fast {
				return chatConfig.FastChatModel, nil
			}
			return chatConfig.ChatModel, nil
		},
	)
}
