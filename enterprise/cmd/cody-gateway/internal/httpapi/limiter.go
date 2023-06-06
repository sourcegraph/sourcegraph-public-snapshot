package httpapi

import (
	"net/http"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/response"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/types"
	sgtrace "github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func rateLimit(
	baseLogger log.Logger,
	eventLogger events.Logger,
	cache limiter.RedisStore,
	concurrencyLimitConfig codygateway.ActorConcurrencyLimitConfig,
	next http.Handler,
) http.Handler {
	baseLogger = baseLogger.Scoped("rateLimit", "rate limit handler")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		act := actor.FromContext(r.Context())
		logger := act.Logger(sgtrace.Logger(r.Context(), baseLogger))

		feature, err := extractFeature(r)
		if err != nil {
			response.JSONError(logger, w, http.StatusBadRequest, err)
			return
		}

		l, ok := act.Limiter(baseLogger, cache, feature, concurrencyLimitConfig)
		if !ok {
			response.JSONError(logger, w, http.StatusForbidden, errors.Newf("no access to feature %s", feature))
			return
		}

		commit, err := l.TryAcquire(r.Context())
		if err != nil {
			limitedCause := "quota"
			var concurrencyLimitExceeded actor.ErrConcurrencyLimitExceeded
			if errors.As(err, &concurrencyLimitExceeded) {
				limitedCause = "concurrency"
			}

			if loggerErr := eventLogger.LogEvent(
				r.Context(),
				events.Event{
					Name:       codygateway.EventNameRateLimited,
					Source:     act.Source.Name(),
					Identifier: act.ID,
					Metadata: map[string]any{
						"error": err.Error(),
						codygateway.CompletionsEventFeatureMetadataField: feature,
						"cause": limitedCause,
					},
				},
			); loggerErr != nil {
				logger.Error("failed to log event", log.Error(loggerErr))
			}

			var rateLimitExceeded limiter.RateLimitExceededError
			if errors.As(err, &rateLimitExceeded) {
				rateLimitExceeded.WriteResponse(w)
				return
			}

			if errors.Is(err, limiter.NoAccessError{}) {
				response.JSONError(logger, w, http.StatusForbidden, err)
				return
			}

			if limitedCause == "concurrency" {
				concurrencyLimitExceeded.WriteResponse(w)
				return
			}

			response.JSONError(logger, w, http.StatusInternalServerError, err)
			return
		}

		responseRecorder := response.NewStatusHeaderRecorder(w)
		next.ServeHTTP(responseRecorder, r)

		// If response is healthy, consume the rate limit
		if responseRecorder.StatusCode >= 200 || responseRecorder.StatusCode < 300 {
			if err := commit(); err != nil {
				logger.Error("failed to commit rate limit consumption", log.Error(err))
			}
		}
	})
}

func extractFeature(r *http.Request) (types.CompletionsFeature, error) {
	h := strings.TrimSpace(r.Header.Get(codygateway.FeatureHeaderName))
	if h == "" {
		return "", errors.Newf("%s header is required", codygateway.FeatureHeaderName)
	}
	feature := types.CompletionsFeature(h)
	if !feature.IsValid() {
		return "", errors.Newf("invalid value for %s", codygateway.FeatureHeaderName)
	}
	return feature, nil
}
