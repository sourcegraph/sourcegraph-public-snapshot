package httpapi

import (
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var autocompleteConfig *conftypes.AutocompleteConfig

// NewCodeCompletionsHandler is an http handler which sends back code completion results.
func NewCodeCompletionsHandler(logger log.Logger, db database.DB) http.Handler {
	logger = logger.Scoped("code", "code completions handler")
	rl := NewRateLimiter(db, redispool.Store, types.CompletionsFeatureCode)
	autocompleteConfig = conf.GetAutocompleteConfig(conf.Get().SiteConfig())
	go conf.Watch(func() {
		oldProvider := autocompleteConfig.ProviderName()
		autocompleteConfig = conf.GetAutocompleteConfig(conf.Get().SiteConfig())
		logger.Info("Updating autocomplete config", log.String("Old Provider", string(oldProvider)), log.String("New Provider", string(autocompleteConfig.ProviderName())))
	})

	return newCompletionsHandler(
		logger,
		types.CompletionsFeatureCode,
		rl,
		"code",
		autocompleteConfig,
		func(requestParams types.CodyCompletionRequestParameters) (string, error) {
			if autocompleteConfig == nil {
				return "", errors.New("autocomplete not configured or disabled")
			}
			if isAllowedCustomModel(requestParams.Model) {
				return requestParams.Model, nil
			}
			if requestParams.Model != "" {
				return "", errors.New("Unsupported custom model")
			}
			return autocompleteConfig.Model, nil
		},
	)
}

// We only allow dotcom clients to select a custom code model and maintain an allowlist for which
// custom values we support
func isAllowedCustomModel(model string) bool {
	if !(envvar.SourcegraphDotComMode()) {
		return false
	}

	switch model {
	case "fireworks/accounts/fireworks/models/starcoder-16b-w8a16":
		fallthrough
	case "fireworks/accounts/fireworks/models/starcoder-7b-w8a16":
		fallthrough
	case "fireworks/accounts/fireworks/models/starcoder-3b-w8a16":
		fallthrough
	case "fireworks/accounts/fireworks/models/starcoder-1b-w8a16":
		fallthrough
	case "fireworks/accounts/fireworks/models/llama-v2-7b-code":
		fallthrough
	case "fireworks/accounts/fireworks/models/llama-v2-13b-code":
		fallthrough
	case "fireworks/accounts/fireworks/models/llama-v2-13b-code-instruct":
		fallthrough
	case "fireworks/accounts/fireworks/models/wizardcoder-15b":
		return true
	}

	return false
}
