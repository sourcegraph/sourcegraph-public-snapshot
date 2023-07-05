package httpapi

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/httpapi/completions"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/httpapi/embeddings"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/httpapi/featurelimiter"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/httpapi/requestlogger"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/notify"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/instrumentation"
)

type Config struct {
	RateLimitNotifier          notify.RateLimitNotifier
	AnthropicAccessToken       string
	AnthropicAllowedModels     []string
	AnthropicMaxTokensToSample int
	OpenAIAccessToken          string
	OpenAIOrgID                string
	OpenAIAllowedModels        []string
	EmbeddingsAllowedModels    []string
}

func NewHandler(logger log.Logger, eventLogger events.Logger, rs limiter.RedisStore, authr *auth.Authenticator, config *Config) http.Handler {
	// Add a prefix to the store for globally unique keys and simpler pruning.
	rs = limiter.NewPrefixRedisStore("rate_limit:", rs)
	r := mux.NewRouter()

	// V1 service routes
	v1router := r.PathPrefix("/v1").Subrouter()

	if config.AnthropicAccessToken != "" {
		v1router.Path("/completions/anthropic").Methods(http.MethodPost).Handler(
			instrumentation.HTTPMiddleware("v1.completions.anthropic",
				authr.Middleware(
					requestlogger.Middleware(
						logger,
						completions.NewAnthropicHandler(
							logger,
							eventLogger,
							rs,
							config.RateLimitNotifier,
							config.AnthropicAccessToken,
							config.AnthropicAllowedModels,
							config.AnthropicMaxTokensToSample,
						),
					),
				),
				otelhttp.WithPublicEndpoint(),
			),
		)
	}
	if config.OpenAIAccessToken != "" {
		v1router.Path("/completions/openai").Methods(http.MethodPost).Handler(
			instrumentation.HTTPMiddleware("v1.completions.openai",
				authr.Middleware(
					requestlogger.Middleware(
						logger,
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
				otelhttp.WithPublicEndpoint(),
			),
		)

		v1router.Path("/embeddings/models").Methods(http.MethodGet).Handler(
			instrumentation.HTTPMiddleware("v1.embeddings.models",
				authr.Middleware(
					requestlogger.Middleware(
						logger,
						embeddings.NewListHandler(),
					),
				),
				otelhttp.WithPublicEndpoint(),
			),
		)

		v1router.Path("/embeddings").Methods(http.MethodPost).Handler(
			instrumentation.HTTPMiddleware("v1.embeddings",
				authr.Middleware(
					requestlogger.Middleware(
						logger,
						embeddings.NewHandler(
							logger,
							eventLogger,
							rs,
							config.RateLimitNotifier,
							embeddings.ModelFactoryMap{
								embeddings.ModelNameOpenAIAda: embeddings.NewOpenAIClient(httpcli.ExternalClient, config.OpenAIAccessToken),
							},
							config.EmbeddingsAllowedModels,
						),
					),
				),
				otelhttp.WithPublicEndpoint(),
			),
		)
	}

	// Register a route where actors can retrieve their current rate limit state.
	v1router.Path("/limits").Methods(http.MethodGet).Handler(
		instrumentation.HTTPMiddleware("v1.limits",
			authr.Middleware(
				requestlogger.Middleware(
					logger,
					featurelimiter.ListLimitsHandler(logger, eventLogger, rs),
				),
			),
			otelhttp.WithPublicEndpoint(),
		),
	)

	return r
}
