package embeddings

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/log"
	"golang.org/x/exp/slices"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/httpapi/featurelimiter"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/notify"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/response"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	sgtrace "github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const usageHeaderName = "X-Token-Usage"

func NewHandler(
	baseLogger log.Logger,
	eventLogger events.Logger,
	rs limiter.RedisStore,
	rateLimitNotifier notify.RateLimitNotifier,
	mf ModelFactory,
	allowedModels []string,
) http.Handler {
	baseLogger = baseLogger.Scoped("embeddingshandler", "The HTTP API handler for the embeddings endpoint.")

	return featurelimiter.HandleFeature(
		baseLogger,
		eventLogger,
		rs,
		rateLimitNotifier,
		codygateway.FeatureEmbeddings,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			act := actor.FromContext(r.Context())
			logger := act.Logger(sgtrace.Logger(r.Context(), baseLogger))

			// This will never be nil as the rate limiter middleware checks this before.
			// TODO: Should we read the rate limit from context, and store it in the rate
			// limiter to make this less dependent on these two logics to remain the same?
			rateLimit, ok := act.RateLimits[codygateway.FeatureEmbeddings]
			if !ok {
				response.JSONError(logger, w, http.StatusInternalServerError, errors.Newf("rate limit for %q not found", string(codygateway.FeatureEmbeddings)))
				return
			}

			// Parse the request body.
			var body codygateway.EmbeddingsRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				response.JSONError(logger, w, http.StatusBadRequest, errors.Wrap(err, "failed to parse request body"))
				return
			}

			if !isAllowedModel(intersection(allowedModels, rateLimit.AllowedModels), body.Model) {
				response.JSONError(logger, w, http.StatusBadRequest, errors.Newf("model %q is not allowed", body.Model))
				return
			}

			c, ok := mf.ForModel(body.Model)
			if !ok {
				response.JSONError(logger, w, http.StatusBadRequest, errors.Newf("model %q not known", body.Model))
				return
			}

			// Add the client type to the logger fields.
			logger = logger.With(log.String("client", fmt.Sprintf("%T", c)))

			upstreamStarted := time.Now()
			var (
				upstreamFinished time.Duration
				// resolvedStatusCode is the status code that we returned to the
				// client - in most case it is the same as upstreamStatusCode,
				// but sometimes we write something different.
				resolvedStatusCode int = -1
				usedTokens         int = -1
			)
			defer func() {
				err := eventLogger.LogEvent(
					r.Context(),
					events.Event{
						Name:       codygateway.EventNameEmbeddingsFinished,
						Source:     act.Source.Name(),
						Identifier: act.ID,
						Metadata: map[string]any{
							"model": body.Model,
							codygateway.CompletionsEventFeatureMetadataField: codygateway.CompletionsEventFeatureEmbeddings,
							"upstream_request_duration_ms":                   upstreamFinished.Milliseconds(),
							"resolved_status_code":                           resolvedStatusCode,
							codygateway.EmbeddingsTokenUsageMetadataField:    usedTokens,
						},
					},
				)
				if err != nil {
					logger.Error("failed to log event", log.Error(err))
				}
			}()

			resp, ut, err := c.GenerateEmbeddings(r.Context(), body)
			usedTokens = ut
			upstreamFinished = time.Since(upstreamStarted)
			if err != nil {
				// If a status error is returned, pass through the code and error
				var statusCodeErr response.HTTPStatusCodeError
				if errors.As(err, &statusCodeErr) {
					resolvedStatusCode = statusCodeErr.HTTPStatusCode()
					response.JSONError(logger, w, statusCodeErr.HTTPStatusCode(), statusCodeErr)
					return
				}

				// Return generic error for other unexpected errors.
				resolvedStatusCode = http.StatusInternalServerError
				response.JSONError(logger, w, http.StatusInternalServerError, err)
				return
			}

			w.Header().Add(usageHeaderName, strconv.Itoa(usedTokens))

			data, err := json.Marshal(resp)
			if err != nil {
				resolvedStatusCode = http.StatusInternalServerError
				response.JSONError(logger, w, http.StatusInternalServerError, errors.Wrap(err, "failed to marshal response"))
				return
			}

			w.Header().Add("Content-Type", "application/json; charset=utf-8")
			_, _ = w.Write(data)
		}))
}

func isAllowedModel(allowedModels []string, model string) bool {
	for _, m := range allowedModels {
		if strings.EqualFold(m, model) {
			return true
		}
	}
	return false
}

func intersection(a, b []string) (c []string) {
	for _, val := range a {
		if slices.Contains(b, val) {
			c = append(c, val)
		}
	}
	return c
}
