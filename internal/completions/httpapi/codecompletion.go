package httpapi

import (
	"context"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
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
		db.Users(),
		telemetryrecorder.New(db),
		types.CompletionsFeatureCode,
		rl,
		"code",
		func(ctx context.Context, requestParams types.CodyCompletionRequestParameters, c *conftypes.CompletionsConfig) (string, error) {
			customModel := allowedCustomModel(ctx, requestParams.Model)
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

func allowedCustomModel(ctx context.Context, model string) string {
	switch model {
	// These special model strings allow the server to choose the model. This allows us to instantly
	// route traffic from Fireworks multi-tenant cluster to our single-tenant cluster and
	// vice-versa, without the client having to know about it
	case "fireworks/starcoder-16b",
		"fireworks/starcoder-7b":

		flags := featureflag.FromContext(ctx)
		singleTenant := flags.GetBoolOr("cody-autocomplete-default-starcoder-hybrid-sourcegraph", false)

		if model == "fireworks/starcoder-16b" {
			if singleTenant {
				return "fireworks/accounts/sourcegraph/models/starcoder-16b"
			}
			return "fireworks/accounts/fireworks/models/starcoder-16b-w8a16"
		}

		if singleTenant {
			return "fireworks/accounts/sourcegraph/models/starcoder-7b"
		}
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
		"fireworks/accounts/fireworks/models/wizardcoder-15b",
		"anthropic/claude-instant-1.2-cyan",
		"anthropic/claude-instant-1.2":
		return model
	}

	return ""
}
