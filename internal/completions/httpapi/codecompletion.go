package httpapi

import (
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

// NewCodeCompletionsHandler is an http handler which sends back code completion results.
func NewCodeCompletionsHandler(logger log.Logger, db database.DB) http.Handler {
	logger = logger.Scoped("code", "code completions handler")
	rl := NewRateLimiter(db, redispool.Store, types.CompletionsFeatureCode)
	return newCompletionsHandler(
		logger,
		types.CompletionsFeatureCode,
		rl,
		"code",
		func(requestParams types.CodyCompletionRequestParameters, c *conftypes.CompletionsConfig) string {
			customModel := allowedCustomModel(requestParams.Model)
			if customModel == "" {
				return c.CompletionModel
			}
			return customModel
		},
	)
}

// We only allow dotcom clients to select a custom code model and maintain an allowlist for which
// custom values we support
func allowedCustomModel(model string) string {
	if !envvar.SourcegraphDotComMode() {
		return ""
	}

	switch model {
	case "fireworks/accounts/fireworks/models/starcoder-16b-w8a16":
	case "fireworks/accounts/fireworks/models/starcoder-7b-w8a16":
	case "fireworks/accounts/fireworks/models/starcoder-3b-w8a16":
	case "fireworks/accounts/fireworks/models/starcoder-1b-w8a16":
		return model
	}

	return ""
}
