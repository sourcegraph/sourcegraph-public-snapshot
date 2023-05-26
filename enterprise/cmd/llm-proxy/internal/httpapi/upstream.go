package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sourcegraph/log"
	"golang.org/x/exp/slices"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/actor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/events"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/limiter"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/response"
	llmproxy "github.com/sourcegraph/sourcegraph/enterprise/internal/llm-proxy"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	sgtrace "github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type bodyTransformer[T any] func(*T)
type requestTransformer func(*http.Request)
type requestMetadataRetriever[T any] func(T) (promptCharacterCount int, model string, additionalMetadata map[string]any)
type responseParser[T any] func(T, io.Reader) (completionCharacterCount int)

func makeUpstreamHandler[ReqT any](
	baseLogger log.Logger,
	eventLogger events.Logger,
	rs limiter.RedisStore,
	upstreamName, upstreamAPIURL string,
	allowedModels []string,
	bodyTrans bodyTransformer[ReqT],
	rmr requestMetadataRetriever[ReqT],
	reqTrans requestTransformer,
	respParser responseParser[ReqT],
) http.Handler {
	baseLogger = baseLogger.Scoped(strings.ToLower(upstreamName), fmt.Sprintf("%s upstream handler", upstreamName))

	return rateLimit(baseLogger, eventLogger, limiter.NewPrefixRedisStore("rate_limit:", rs), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		act := actor.FromContext(r.Context())
		logger := act.Logger(sgtrace.Logger(r.Context(), baseLogger))

		feature, err := extractFeature(r)
		if err != nil {
			response.JSONError(logger, w, http.StatusBadRequest, err)
			return
		}

		// This will never be nil as the rate limiter middleware checks this before.
		// TODO: Should we read the rate limit from context, and store it in the rate
		// limiter to make this less dependent on these two logics to remain the same?
		rateLimit, ok := act.RateLimits[feature]
		if !ok {
			response.JSONError(logger, w, http.StatusInternalServerError, errors.Wrapf(err, "rate limit for %q not found", string(feature)))
			return
		}

		// Parse the request body.
		var body ReqT
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			response.JSONError(logger, w, http.StatusBadRequest, errors.Wrap(err, "failed to parse request body"))
			return
		}

		bodyTrans(&body)

		// Re-marshal the payload for upstream to unset metadata and remove any properties
		// not known to us.
		upstreamPayload, err := json.Marshal(body)
		if err != nil {
			response.JSONError(logger, w, http.StatusInternalServerError, errors.Wrap(err, "failed to marshal request body"))
			return
		}

		// Create a new request to send upstream, making sure we retain the same context.
		req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, upstreamAPIURL, bytes.NewReader(upstreamPayload))
		if err != nil {
			response.JSONError(logger, w, http.StatusInternalServerError, errors.Wrap(err, "failed to create request"))
			return
		}

		// Run the request transformer.
		reqTrans(req)

		promptCharacterCount, model, am := rmr(body)

		if !isAllowedModel(intersection(allowedModels, rateLimit.AllowedModels), model) {
			response.JSONError(logger, w, http.StatusBadRequest, errors.Newf("model %q is not allowed", model))
			return
		}

		{
			metadata := map[string]any{}
			for k, v := range am {
				metadata[k] = v
			}
			metadata["prompt_character_count"] = promptCharacterCount
			metadata["model"] = model
			metadata[llmproxy.CompletionsEventFeatureMetadataField] = feature
			err = eventLogger.LogEvent(
				r.Context(),
				events.Event{
					Name:       llmproxy.EventNameCompletionsStarted,
					Source:     act.Source.Name(),
					Identifier: act.ID,
					Metadata:   metadata,
				},
			)
			if err != nil {
				logger.Error("failed to log event", log.Error(err))
			}
		}

		upstreamStarted := time.Now()
		var (
			upstreamStatusCode       int = -1
			completionCharacterCount int = -1
		)
		defer func() {
			err := eventLogger.LogEvent(
				r.Context(),
				events.Event{
					Name:       llmproxy.EventNameCompletionsFinished,
					Source:     act.Source.Name(),
					Identifier: act.ID,
					Metadata: map[string]any{
						llmproxy.CompletionsEventFeatureMetadataField: feature,
						"upstream_request_duration_ms":                time.Since(upstreamStarted).Milliseconds(),
						"upstream_status_code":                        upstreamStatusCode,
						"completion_character_count":                  completionCharacterCount,
					},
				},
			)
			if err != nil {
				logger.Error("failed to log event", log.Error(err))
			}
		}()

		resp, err := httpcli.ExternalDoer.Do(req)
		if err != nil {
			response.JSONError(logger, w, http.StatusInternalServerError,
				errors.Wrapf(err, "failed to make request to upstream provider %s", upstreamName))
			return
		}
		defer func() { _ = resp.Body.Close() }()

		// Forward upstream http headers.
		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}

		// Forward status code.
		upstreamStatusCode = resp.StatusCode
		w.WriteHeader(resp.StatusCode)

		// Set up a buffer to capture the response as it's streamed and sent to the client.
		var responseBuf bytes.Buffer
		respBody := io.TeeReader(resp.Body, &responseBuf)
		// Forward response to client.
		_, _ = io.Copy(w, respBody)

		if upstreamStatusCode >= 200 && upstreamStatusCode < 300 {
			// Pass reader to response transformer to capture token counts.
			completionCharacterCount = respParser(body, &responseBuf)

		} else if upstreamStatusCode >= 500 {
			logger.Error("error from upstream",
				log.String("url", upstreamAPIURL),
				log.Int("status_code", upstreamStatusCode))
		}
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
