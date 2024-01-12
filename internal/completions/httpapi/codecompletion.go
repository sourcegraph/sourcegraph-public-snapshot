package httpapi

import (
	"context"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/telemetryrecorder"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewCodeCompletionsHandler is an http handler which sends back code completion results.
func NewCodeCompletionsHandler(logger log.Logger, db database.DB) http.Handler {
	logger = logger.Scoped("code")
	rl := NewRateLimiter(db, redispool.Store, types.CompletionsFeatureCode)
	return newCompletionsHandler(
		logger,
		db,
		db.Users(),
		db.AccessTokens(),
		telemetryrecorder.New(db),
		types.CompletionsFeatureCode,
		rl,
		"code",
		func(_ context.Context, requestParams types.CodyCompletionRequestParameters, c *conftypes.CompletionsConfig) (string, error) {
			customModel := allowedCustomModel(requestParams.Model)
			if customModel != "" {
				return customModel, nil
			}
			if requestParams.Model != "" {
				return "", errors.Newf("Unsupported code completion model %q", requestParams.Model)
			}
			return c.CompletionModel, nil
		},
	)
}

func allowedCustomModel(model string) string {
	switch model {
	// These virtual model strings allow the server to choose the model.
	// TODO: Move the translation of these virtual model strings to cody gateway.
	case "fireworks/starcoder-16b":
		return "fireworks/accounts/fireworks/models/starcoder-16b-w8a16"
	case "fireworks/starcoder-7b":
		return "fireworks/accounts/fireworks/models/starcoder-7b-w8a16"

	case "fireworks/accounts/fireworks/models/starcoder-16b-w8a16",
		"fireworks/accounts/fireworks/models/starcoder-7b-w8a16",
		"fireworks/accounts/fireworks/models/starcoder-3b-w8a16",
		"fireworks/accounts/fireworks/models/starcoder-1b-w8a16",
		"fireworks/accounts/sourcegraph/models/starcoder-7b",
		"fireworks/accounts/sourcegraph/models/starcoder-16b",
		"fireworks/accounts/fireworks/models/llama-v2-7b-code",
		"fireworks/accounts/fireworks/models/llama-v2-13b-code",
		"fireworks/accounts/fireworks/models/llama-v2-13b-code-instruct",
		"fireworks/accounts/fireworks/models/llama-v2-34b-code-instruct",
		"fireworks/accounts/fireworks/models/mistral-7b-instruct-4k",
		"anthropic/claude-instant-1.2-cyan",
		"anthropic/claude-instant-1.2",
		"anthropic/claude-instant-v1",
		"anthropic/claude-instant-1":
		return model
	}

	return ""
}
