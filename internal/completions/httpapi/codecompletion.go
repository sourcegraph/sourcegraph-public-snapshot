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

// NewCodeCompletionsHandler is an http handler which sends back code completion results.
func NewCodeCompletionsHandler(logger log.Logger, db database.DB) http.Handler {
	logger = logger.Scoped("code", "code completions handler")
	rl := NewRateLimiter(db, redispool.Store, types.CompletionsFeatureCode)
	getConfig := func() conftypes.ProviderConfig {
		return conf.GetAutocompleteConfig(conf.Get().SiteConfig())
	}

	return newCompletionsHandler(
		logger,
		types.CompletionsFeatureCode,
		rl,
		"code",
		getConfig,
		func(requestParams types.CodyCompletionRequestParameters) (string, error) {
			config := conf.GetAutocompleteConfig(conf.Get().SiteConfig())
			if config == nil {
				return "", errors.New("completions are not configured or disabled")
			}
			if isAllowedCustomModel(requestParams.Model) {
				return requestParams.Model, nil
			}
			if requestParams.Model != "" {
				return "", errors.New("Unsupported custom model")
			}
			return config.Model, nil
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
