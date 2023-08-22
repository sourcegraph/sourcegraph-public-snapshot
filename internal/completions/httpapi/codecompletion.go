package httpapi

import (
	"net/http"

	"github.com/sourcegraph/log"

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
			if requestParams.Model == "" {
				return c.CompletionModel
			}
			return requestParams.Model
		},
	)
}
