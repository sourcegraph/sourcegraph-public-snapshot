package httpapi

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/httpapi/completions"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/httpapi/embeddings"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/notify"
	"github.com/sourcegraph/sourcegraph/internal/instrumentation"
)

type Config struct {
	RateLimitNotifier       notify.RateLimitNotifier
	AnthropicAccessToken    string
	AnthropicAllowedModels  []string
	OpenAIAccessToken       string
	OpenAIOrgID             string
	OpenAIAllowedModels     []string
	EmbeddingsAllowedModels []string
}

func NewHandler(logger log.Logger, eventLogger events.Logger, rs limiter.RedisStore, authr *auth.Authenticator, config *Config) http.Handler {
	r := mux.NewRouter()

	// V1 service routes
	v1router := r.PathPrefix("/v1").Subrouter()

	if config.AnthropicAccessToken != "" {
		v1router.Path("/completions/anthropic").Methods(http.MethodPost).Handler(
			instrumentation.HTTPMiddleware("v1.completions.anthropic",
				authr.Middleware(
					completions.NewAnthropicHandler(
						logger,
						eventLogger,
						rs,
						config.RateLimitNotifier,
						config.AnthropicAccessToken,
						config.AnthropicAllowedModels,
					),
				),
			),
		)
	}
	if config.OpenAIAccessToken != "" {
		v1router.Path("/completions/openai").Methods(http.MethodPost).Handler(
			instrumentation.HTTPMiddleware("v1.completions.openai",
				authr.Middleware(
					completions.NewOpenAIHandler(
						logger,
						eventLogger,
						rs,
						config.RateLimitNotifier,
						config.OpenAIAccessToken,
						config.OpenAIOrgID,
						config.OpenAIAllowedModels,
					),
				),
			),
		)

		v1router.Path("/embeddings/models").Methods(http.MethodGet).Handler(
			instrumentation.HTTPMiddleware("v1.embeddings.models",
				authr.Middleware(
					embeddings.NewListHandler(),
				),
			),
		)

		v1router.Path("/embeddings").Methods(http.MethodPost).Handler(
			instrumentation.HTTPMiddleware("v1.embeddings",
				authr.Middleware(
					embeddings.NewHandler(
						logger,
						eventLogger,
						rs,
						config.RateLimitNotifier,
						embeddings.ModelFactoryMap{
							embeddings.ModelNameOpenAIAda: embeddings.NewOpenAIClient(config.OpenAIAccessToken),
						},
						config.EmbeddingsAllowedModels,
					),
				),
			),
		)
	}

	return r
}
