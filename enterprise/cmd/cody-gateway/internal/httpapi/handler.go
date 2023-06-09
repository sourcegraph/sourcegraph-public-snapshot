package httpapi

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/httpapi/completions"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/httpapi/embeddings"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codygateway"
)

type Config struct {
	RateLimitAlerter        func(actor *actor.Actor, feature codygateway.Feature, usagePercentage float32, ttl time.Duration)
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
			authr.Middleware(
				completions.NewAnthropicHandler(
					logger,
					eventLogger,
					rs,
					config.RateLimitAlerter,
					config.AnthropicAccessToken,
					config.AnthropicAllowedModels,
				),
			),
		)
	}
	if config.OpenAIAccessToken != "" {
		v1router.Path("/completions/openai").Methods(http.MethodPost).Handler(
			authr.Middleware(
				completions.NewOpenAIHandler(
					logger,
					eventLogger,
					rs,
					config.RateLimitAlerter,
					config.OpenAIAccessToken,
					config.OpenAIOrgID,
					config.OpenAIAllowedModels,
				),
			),
		)

		v1router.Path("/embeddings/models").Methods(http.MethodGet).Handler(
			authr.Middleware(
				embeddings.NewListHandler(),
			),
		)

		v1router.Path("/embeddings").Methods(http.MethodPost).Handler(
			authr.Middleware(
				embeddings.NewHandler(
					logger,
					eventLogger,
					rs,
					config.RateLimitAlerter,
					embeddings.ModelFactoryMap{
						embeddings.ModelNameOpenAIAda: embeddings.NewOpenAIClient(config.OpenAIAccessToken),
					},
					config.EmbeddingsAllowedModels,
				),
			),
		)
	}

	return r
}
