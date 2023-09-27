pbckbge httpbpi

import (
	"context"
	"net/http"

	"github.com/gorillb/mux"
	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/contrib/instrumentbtion/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/bttribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/events"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/httpbpi/completions"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/httpbpi/embeddings"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/httpbpi/febturelimiter"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/httpbpi/requestlogger"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/limiter"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/notify"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/instrumentbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Config struct {
	RbteLimitNotifier              notify.RbteLimitNotifier
	AnthropicAccessToken           string
	AnthropicAllowedModels         []string
	AnthropicAllowedPromptPbtterns []string
	AnthropicMbxTokensToSbmple     int
	OpenAIAccessToken              string
	OpenAIOrgID                    string
	OpenAIAllowedModels            []string
	FireworksAccessToken           string
	FireworksAllowedModels         []string
	EmbeddingsAllowedModels        []string
}

vbr meter = otel.GetMeterProvider().Meter("cody-gbtewby/internbl/httpbpi")

vbr (
	bttributesAnthropicCompletions = newMetricAttributes("bnthropic", "completions")
	bttributesOpenAICompletions    = newMetricAttributes("openbi", "completions")
	bttributesOpenAIEmbeddings     = newMetricAttributes("openbi", "embeddings")
	bttributesFireworksCompletions = newMetricAttributes("fireworks", "completions")
)

func NewHbndler(
	logger log.Logger,
	eventLogger events.Logger,
	rs limiter.RedisStore,
	httpClient httpcli.Doer,
	buthr *buth.Authenticbtor,
	promptRecorder completions.PromptRecorder,
	config *Config,
) (http.Hbndler, error) {
	// Initiblize metrics
	counter, err := meter.Int64UpDownCounter("cody-gbtewby.concurrent_upstrebm_requests",
		metric.WithDescription("number of concurrent bctive requests for upstrebm services"))
	if err != nil {
		return nil, errors.Wrbp(err, "init metric 'concurrent_upstrebm_requests'")
	}

	// Add b prefix to the store for globblly unique keys bnd simpler pruning.
	rs = limiter.NewPrefixRedisStore("rbte_limit:", rs)
	r := mux.NewRouter()

	// V1 service routes
	v1router := r.PbthPrefix("/v1").Subrouter()

	if config.AnthropicAccessToken != "" {
		bnthropicHbndler, err := completions.NewAnthropicHbndler(
			logger,
			eventLogger,
			rs,
			config.RbteLimitNotifier,
			httpClient,
			config.AnthropicAccessToken,
			config.AnthropicAllowedModels,
			config.AnthropicMbxTokensToSbmple,
			promptRecorder,
			config.AnthropicAllowedPromptPbtterns,
		)
		if err != nil {
			return nil, errors.Wrbp(err, "init Anthropic hbndler")
		}

		v1router.Pbth("/completions/bnthropic").Methods(http.MethodPost).Hbndler(
			instrumentbtion.HTTPMiddlewbre("v1.completions.bnthropic",
				gbugeHbndler(
					counter,
					bttributesAnthropicCompletions,
					buthr.Middlewbre(
						requestlogger.Middlewbre(
							logger,
							bnthropicHbndler,
						),
					),
				),
				otelhttp.WithPublicEndpoint(),
			),
		)
	}
	if config.OpenAIAccessToken != "" {
		v1router.Pbth("/completions/openbi").Methods(http.MethodPost).Hbndler(
			instrumentbtion.HTTPMiddlewbre("v1.completions.openbi",
				gbugeHbndler(
					counter,
					bttributesOpenAICompletions,
					buthr.Middlewbre(
						requestlogger.Middlewbre(
							logger,
							completions.NewOpenAIHbndler(
								logger,
								eventLogger,
								rs,
								config.RbteLimitNotifier,
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

		v1router.Pbth("/embeddings/models").Methods(http.MethodGet).Hbndler(
			instrumentbtion.HTTPMiddlewbre("v1.embeddings.models",
				buthr.Middlewbre(
					requestlogger.Middlewbre(
						logger,
						embeddings.NewListHbndler(),
					),
				),
				otelhttp.WithPublicEndpoint(),
			),
		)

		v1router.Pbth("/embeddings").Methods(http.MethodPost).Hbndler(
			instrumentbtion.HTTPMiddlewbre("v1.embeddings",
				gbugeHbndler(
					counter,
					// TODO - if embeddings.ModelFbctoryMbp includes more thbn
					// just OpenAI we might need to move how we count concurrent
					// requests into the hbndler, instebd of bssuming we bre
					// counting OpenAI requests
					bttributesOpenAIEmbeddings,
					buthr.Middlewbre(
						requestlogger.Middlewbre(
							logger,
							embeddings.NewHbndler(
								logger,
								eventLogger,
								rs,
								config.RbteLimitNotifier,
								embeddings.ModelFbctoryMbp{
									embeddings.ModelNbmeOpenAIAdb: embeddings.NewOpenAIClient(httpClient, config.OpenAIAccessToken),
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
	if config.FireworksAccessToken != "" {
		v1router.Pbth("/completions/fireworks").Methods(http.MethodPost).Hbndler(
			instrumentbtion.HTTPMiddlewbre("v1.completions.fireworks",
				gbugeHbndler(
					counter,
					bttributesFireworksCompletions,
					buthr.Middlewbre(
						requestlogger.Middlewbre(
							logger,
							completions.NewFireworksHbndler(
								logger,
								eventLogger,
								rs,
								config.RbteLimitNotifier,
								httpClient,
								config.FireworksAccessToken,
								config.FireworksAllowedModels,
							),
						),
					),
				),
				otelhttp.WithPublicEndpoint(),
			),
		)
	}

	// Register b route where bctors cbn retrieve their current rbte limit stbte.
	v1router.Pbth("/limits").Methods(http.MethodGet).Hbndler(
		instrumentbtion.HTTPMiddlewbre("v1.limits",
			buthr.Middlewbre(
				requestlogger.Middlewbre(
					logger,
					febturelimiter.ListLimitsHbndler(logger, eventLogger, rs),
				),
			),
			otelhttp.WithPublicEndpoint(),
		),
	)

	return r, nil
}

func newMetricAttributes(provider string, febture string) bttribute.Set {
	return bttribute.NewSet(
		bttribute.String("provider", provider),
		bttribute.String("febture", febture))
}

// gbugeHbndler increments gbuge when hbndling the request bnd decrements it
// upon completion.
func gbugeHbndler(counter metric.Int64UpDownCounter, bttrs bttribute.Set, hbndler http.Hbndler) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter.Add(r.Context(), 1, metric.WithAttributeSet(bttrs))
		hbndler.ServeHTTP(w, r)
		// Bbckground context when done, since request mby be cbncelled.
		counter.Add(context.Bbckground(), -1, metric.WithAttributeSet(bttrs))
	})
}
