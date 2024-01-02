package featurelimiter

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/notify"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/response"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	sgtrace "github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type contextKey string

const contextKeyFeature contextKey = "feature"

// GetFeature gets the feature used by Handle or HandleFeature.
func GetFeature(ctx context.Context) codygateway.Feature {
	if f, ok := ctx.Value(contextKeyFeature).(codygateway.Feature); ok {
		return f
	}
	return ""
}

// Handle extracts features from codygateway.FeatureHeaderName and uses it to
// determine the appropriate per-feature rate limits applied for an actor. It
// only limits per-request.
func Handle(
	baseLogger log.Logger,
	eventLogger events.Logger,
	cache limiter.RedisStore,
	rateLimitNotifier notify.RateLimitNotifier,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		feature, err := extractFeature(r)
		if err != nil {
			response.JSONError(baseLogger, w, http.StatusBadRequest, err)
			return
		}

		extractUsage := func(responseHeaders http.Header) (int, error) {
			return 1, nil // per-request, so always consume 1 rate limit
		}

		HandleFeature(baseLogger, eventLogger, cache, rateLimitNotifier, feature, next, extractUsage).
			ServeHTTP(w, r)
	})
}

func extractFeature(r *http.Request) (codygateway.Feature, error) {
	h := strings.TrimSpace(r.Header.Get(codygateway.FeatureHeaderName))
	if h == "" {
		return "", errors.Newf("%s header is required", codygateway.FeatureHeaderName)
	}
	feature := types.CompletionsFeature(h)
	if !feature.IsValid() {
		return "", errors.Newf("invalid value for %s", codygateway.FeatureHeaderName)
	}
	// codygateway.Feature and types.CompletionsFeature map 1:1 for completions.
	return codygateway.Feature(feature), nil
}

// HandleFeature uses a predefined feature to determine the appropriate per-feature
// rate limits applied for an actor.
func HandleFeature(
	baseLogger log.Logger,
	eventLogger events.Logger,
	cache limiter.RedisStore,
	rateLimitNotifier notify.RateLimitNotifier,
	feature codygateway.Feature,
	next http.Handler,
	extractUsage func(responseHeaders http.Header) (int, error),
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		act := actor.FromContext(r.Context())
		logger := act.Logger(sgtrace.Logger(r.Context(), baseLogger))

		r = r.WithContext(context.WithValue(r.Context(), contextKeyFeature, feature))

		l, ok := act.Limiter(logger, cache, feature, rateLimitNotifier)
		if !ok {
			response.JSONError(logger, w, http.StatusForbidden, errors.Newf("no access to feature %s", feature))
			return
		}

		commit, err := l.TryAcquire(r.Context())
		if err != nil {
			limitedCause := "quota"
			defer func() {
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
			}()

			var concurrencyLimitExceeded actor.ErrConcurrencyLimitExceeded
			if errors.As(err, &concurrencyLimitExceeded) {
				limitedCause = "concurrency"
				concurrencyLimitExceeded.WriteResponse(w)
				return
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

			response.JSONError(logger, w, http.StatusInternalServerError, err)
			return
		}

		responseRecorder := response.NewStatusHeaderRecorder(w, logger)
		next.ServeHTTP(responseRecorder, r)

		// If response is healthy, consume the rate limit
		if responseRecorder.StatusCode >= 200 && responseRecorder.StatusCode < 300 {
			usage, err := extractUsage(w.Header())
			if err != nil {
				logger.Error("failed to extract usage", log.Error(err))
			}
			if err := commit(r.Context(), usage); err != nil {
				logger.Error("failed to commit rate limit consumption", log.Error(err))
			}
		}
	})
}

// ListLimitsHandler returns a map of all features and their current rate limit usages.
func ListLimitsHandler(baseLogger log.Logger, redisStore limiter.RedisStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		act := actor.FromContext(r.Context())
		logger := act.Logger(sgtrace.Logger(r.Context(), baseLogger))

		res := map[codygateway.Feature]listLimitElement{}

		// Iterate over all features.
		for _, f := range codygateway.AllFeatures {
			// Get the limiter, but don't log any rate limit events, the only limits enforced
			// here are concurrency limits, and we should not care about those.
			l, ok := act.Limiter(logger, redisStore, f, noopRateLimitNotifier)
			if !ok {
				response.JSONError(logger, w, http.StatusForbidden, errors.Newf("no access to feature %s", f))
				return
			}

			// Capture the current usage.
			currentUsage, expiry, err := l.Usage(r.Context())
			if err != nil {
				if errors.HasType(err, limiter.NoAccessError{}) {
					// No access to this feature, skip.
					continue
				}
				response.JSONError(logger, w, http.StatusInternalServerError, errors.Wrapf(err, "failed to get usage for %s", f))
				return
			}

			// Find the configured rate limit. This should always be set after reading the Usage,
			// but just to be safe, we add an existence check here.
			rateLimit, ok := act.RateLimits[f]
			if !ok {
				response.JSONError(logger, w, http.StatusInternalServerError, errors.Newf("rate limit for %q not found", string(f)))
				return
			}

			el := listLimitElement{
				Limit:         rateLimit.Limit,
				Interval:      rateLimit.Interval.String(),
				Usage:         int64(currentUsage),
				AllowedModels: rateLimit.AllowedModels,
			}
			if !expiry.IsZero() {
				el.Expiry = &expiry
			}
			res[f] = el
		}

		w.Header().Add("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(res); err != nil {
			baseLogger.Debug("failed to marshal json response", log.Error(err))
		}
	})
}

func RefreshLimitsHandler(baseLogger log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		act := actor.FromContext(r.Context())

		if err := act.Update(r.Context()); err != nil {
			logger := act.Logger(trace.Logger(r.Context(), baseLogger))
			if actor.IsErrActorRecentlyUpdated(err) {
				response.JSONError(logger, w, http.StatusTooManyRequests, err)
			} else {
				response.JSONError(logger, w, http.StatusInternalServerError, err)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}

type listLimitElement struct {
	Limit         int64      `json:"limit"`
	Interval      string     `json:"interval"`
	Usage         int64      `json:"usage"`
	Expiry        *time.Time `json:"expiry,omitempty"`
	AllowedModels []string   `json:"allowedModels"`
}

func noopRateLimitNotifier(_ context.Context, _ codygateway.Actor, _ codygateway.Feature, _ float32, _ time.Duration) {
	// nothing
}
