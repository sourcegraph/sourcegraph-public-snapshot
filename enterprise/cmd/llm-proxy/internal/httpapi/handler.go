package httpapi

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/events"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/limiter"
)

type Config struct {
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
				newAnthropicHandler(logger, eventLogger, rs, config.AnthropicAccessToken, config.AnthropicAllowedModels),
			),
		).Methods(http.MethodPost)
	}
	if config.OpenAIAccessToken != "" {
		v1router.Handle(
			"/completions/openai",
			authr.Middleware(
				newOpenAIHandler(logger, eventLogger, rs, config.OpenAIAccessToken, config.OpenAIOrgID, config.OpenAIAllowedModels),
			),
		).Methods(http.MethodPost)
	}

	return r
}
