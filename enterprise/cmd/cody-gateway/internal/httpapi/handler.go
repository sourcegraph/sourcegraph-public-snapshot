package httpapi

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codygateway"
)

type Config struct {
	ConcurrencyLimit       codygateway.ActorConcurrencyLimitConfig
	AnthropicAccessToken   string
	AnthropicAllowedModels []string
	OpenAIAccessToken      string
	OpenAIOrgID            string
	OpenAIAllowedModels    []string
}

func NewHandler(logger log.Logger, eventLogger events.Logger, rs limiter.RedisStore, authr *auth.Authenticator, config *Config) http.Handler {
	r := mux.NewRouter()

	// V1 service routes
	v1router := r.PathPrefix("/v1").Subrouter()

	if config.AnthropicAccessToken != "" {
		v1router.Handle(
			"/completions/anthropic",
			authr.Middleware(
				newAnthropicHandler(logger, eventLogger, rs, config.ConcurrencyLimit, config.AnthropicAccessToken, config.AnthropicAllowedModels),
			),
		).Methods(http.MethodPost)
	}
	if config.OpenAIAccessToken != "" {
		v1router.Handle(
			"/completions/openai",
			authr.Middleware(
				newOpenAIHandler(logger, eventLogger, rs, config.ConcurrencyLimit, config.OpenAIAccessToken, config.OpenAIOrgID, config.OpenAIAllowedModels),
			),
		).Methods(http.MethodPost)
	}

	return r
}
