package httpapi

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/atomic"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/auth"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/httpapi/completions"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/httpapi/embeddings"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/httpapi/featurelimiter"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/httpapi/requestlogger"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/notify"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/instrumentation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

var meter = otel.GetMeterProvider().Meter("cody-gateway/internal/httpapi")

var (
	attributesAnthropicCompletions = newMetricAttributes("anthropic", "completions")
	attributesOpenAICompletions    = newMetricAttributes("openai", "completions")
	attributesOpenAIEmbeddings     = newMetricAttributes("openai", "embeddings")
)

func NewHandler(
	logger log.Logger,
	eventLogger events.Logger,
	rs limiter.RedisStore,
	httpClient httpcli.Doer,
	authr *auth.Authenticator,
	config *Config,
) (http.Handler, error) {
	// Initialize metrics
	anthropicCompletionsRequests := atomic.NewInt64(0)
	openaiCompletionsRequests := atomic.NewInt64(0)
	openaiEmbeddingsRequests := atomic.NewInt64(0)
	if _, err := meter.Int64ObservableGauge("concurrent_upstream_requests",
		metric.WithDescription("number of concurrent active requests for upstream services"),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			o.Observe(anthropicCompletionsRequests.Load(),
				metric.WithAttributeSet(attributesAnthropicCompletions))
			o.Observe(openaiCompletionsRequests.Load(),
				metric.WithAttributeSet(attributesOpenAICompletions))
			o.Observe(openaiEmbeddingsRequests.Load(),
				metric.WithAttributeSet(attributesOpenAIEmbeddings))
			return nil
		})); err != nil {
		return nil, errors.Wrap(err, "init metric concurrent_upstream_requests")
	}

	// Add a prefix to the store for globally unique keys and simpler pruning.
	rs = limiter.NewPrefixRedisStore("rate_limit:", rs)
	r := mux.NewRouter()

	// V1 service routes
	v1router := r.PathPrefix("/v1").Subrouter()

	if config.AnthropicAccessToken != "" {
		v1router.Path("/completions/anthropic").Methods(http.MethodPost).Handler(
			instrumentation.HTTPMiddleware("v1.completions.anthropic",
				gaugeHandler(
					anthropicCompletionsRequests,
					authr.Middleware(
						requestlogger.Middleware(
							logger,
							completions.NewAnthropicHandler(
								logger,
								eventLogger,
								rs,
								config.RateLimitNotifier,
								httpClient,
								config.AnthropicAccessToken,
								config.AnthropicAllowedModels,
								config.AnthropicMaxTokensToSample,
							),
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
				gaugeHandler(
					openaiCompletionsRequests,
					authr.Middleware(
						requestlogger.Middleware(
							logger,
							completions.NewOpenAIHandler(
								logger,
								eventLogger,
								rs,
								config.RateLimitNotifier,
								httpClient,
								config.OpenAIAccessToken,
								config.OpenAIOrgID,
								config.OpenAIAllowedModels,
							),
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
				gaugeHandler(
					// TODO - if embeddings.ModelFactoryMap includes more than
					// just OpenAI we might need to move how we count concurrent
					// requests into the handler
					openaiEmbeddingsRequests,
					authr.Middleware(
						requestlogger.Middleware(
							logger,
							embeddings.NewHandler(
								logger,
								eventLogger,
								rs,
								config.RateLimitNotifier,
								embeddings.ModelFactoryMap{
									embeddings.ModelNameOpenAIAda: embeddings.NewOpenAIClient(httpClient, config.OpenAIAccessToken),
								},
								config.EmbeddingsAllowedModels,
							),
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

	return r, nil
}

func newMetricAttributes(provider string, feature string) attribute.Set {
	return attribute.NewSet(
		attribute.String("provider", provider),
		attribute.String("feature", feature))
}

// gaugeHandler increments gauge when handling the request and decrements it
// upon completion.
func gaugeHandler(gauge *atomic.Int64, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gauge.Inc()
		handler.ServeHTTP(w, r)
		gauge.Dec()
	})
}
