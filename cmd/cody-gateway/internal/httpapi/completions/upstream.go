package completions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/slices"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/httpapi/featurelimiter"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/notify"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/response"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	sgtrace "github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type bodyTransformer[T UpstreamRequest] func(*T, *actor.Actor)
type requestTransformer func(*http.Request)
type requestValidator[T UpstreamRequest] func(T) error
type requestMetadataRetriever[T UpstreamRequest] func(T) (promptCharacterCount int, model string, additionalMetadata map[string]any)
type responseParser[T UpstreamRequest] func(T, io.Reader) (completionCharacterCount int)

// upstreamHandlerMethods declares a set of methods that are used throughout the
// lifecycle of a request to an upstream API. All methods are required.
type upstreamHandlerMethods[ReqT UpstreamRequest] struct {
	// transformBody can be used to modify the request body before it is sent
	// upstream. To manipulate the HTTP request, use transformRequest.
	transformBody bodyTransformer[ReqT]
	// transformRequest can be used to modify the HTTP request before it is sent
	// upstream. To manipulate the body, use transformBody.
	transformRequest requestTransformer
	// getRequestMetadata should extract details about the request we are sending
	// upstream for validation and tracking purposes.
	getRequestMetadata requestMetadataRetriever[ReqT]
	// parseResponse should extract details from the response we get back from
	// upstream for tracking purposes.
	parseResponse responseParser[ReqT]
	// validateRequest can be used to validate the HTTP request before it is sent upstream. Returning a non-nil error will stop further processing and return a 400 HTTP error code
	validateRequest requestValidator[ReqT]
}

type UpstreamRequest interface{}

func makeUpstreamHandler[ReqT UpstreamRequest](
	baseLogger log.Logger,
	eventLogger events.Logger,
	rs limiter.RedisStore,
	rateLimitNotifier notify.RateLimitNotifier,
	httpClient httpcli.Doer,

	// upstreamName is the name of the upstream provider. It MUST match the
	// provider names defined clientside, i.e. "anthropic" or "openai".
	upstreamName string,

	upstreamAPIURL string,
	allowedModels []string,

	methods upstreamHandlerMethods[ReqT],
) http.Handler {
	baseLogger = baseLogger.Scoped(upstreamName, fmt.Sprintf("%s upstream handler", upstreamName)).
		With(log.String("upstream.url", upstreamAPIURL))

	var (
		transformBody      = methods.transformBody
		transformRequest   = methods.transformRequest
		getRequestMetadata = methods.getRequestMetadata
		parseResponse      = methods.parseResponse
	)

	// Convert allowedModels to the Cody Gateway configuration format with the
	// provider as a prefix. This aligns with the models returned when we query
	// for rate limits from actor sources.
	for i := range allowedModels {
		allowedModels[i] = fmt.Sprintf("%s/%s", upstreamName, allowedModels[i])
	}

	return featurelimiter.Handle(
		baseLogger,
		eventLogger,
		rs,
		rateLimitNotifier,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			act := actor.FromContext(r.Context())
			logger := act.Logger(sgtrace.Logger(r.Context(), baseLogger))

			feature := featurelimiter.GetFeature(r.Context())
			if feature == "" {
				response.JSONError(logger, w, http.StatusBadRequest, errors.New("no feature provided"))
				return
			}

			// This will never be nil as the rate limiter middleware checks this before.
			// TODO: Should we read the rate limit from context, and store it in the rate
			// limiter to make this less dependent on these two logics to remain the same?
			rateLimit, ok := act.RateLimits[feature]
			if !ok {
				response.JSONError(logger, w, http.StatusInternalServerError, errors.Newf("rate limit for %q not found", string(feature)))
				return
			}

			// TEMPORARY: Add provider prefixes to AllowedModels for back-compat
			// if it doesn't look like there is a prefix yet.
			//
			// This isn't very robust, but should tide us through a brief transition
			// period until everything deploys and our caches refresh.
			for i := range rateLimit.AllowedModels {
				if !strings.Contains(rateLimit.AllowedModels[i], "/") {
					rateLimit.AllowedModels[i] = fmt.Sprintf("%s/%s", upstreamName, rateLimit.AllowedModels[i])
				}
			}

			// Parse the request body.
			var body ReqT
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				response.JSONError(logger, w, http.StatusBadRequest, errors.Wrap(err, "failed to parse request body"))
				return
			}

			if err := methods.validateRequest(body); err != nil {
				response.JSONError(logger, w, http.StatusBadRequest, errors.Wrap(err, "invalid request"))
				return
			}

			transformBody(&body, act)

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
			transformRequest(req)

			// Retrieve metadata from the initial request.
			promptCharacterCount, model, requestMetadata := getRequestMetadata(body)

			// Match the model against the allowlist of models, which are configured
			// with the Cody Gateway model format "$PROVIDER/$MODEL_NAME". Models
			// are sent as if they were against the upstream API, so they don't have
			// the prefix yet when extracted - we need to add it back here. This
			// full gatewayModel is also used in events tracking.
			gatewayModel := fmt.Sprintf("%s/%s", upstreamName, model)
			if allowed := intersection(allowedModels, rateLimit.AllowedModels); !isAllowedModel(allowed, gatewayModel) {
				response.JSONError(logger, w, http.StatusBadRequest,
					errors.Newf("model %q is not allowed, allowed: [%s]",
						gatewayModel, strings.Join(allowed, ", ")))
				return
			}

			var (
				upstreamStarted        = time.Now()
				upstreamStatusCode int = -1
				// resolvedStatusCode is the status code that we returned to the
				// client - in most case it is the same as upstreamStatusCode,
				// but sometimes we write something different.
				resolvedStatusCode int = -1
				// completionCharacterCount is extracted from parseResponse.
				completionCharacterCount int = -1
			)
			defer func() {
				if span := oteltrace.SpanFromContext(r.Context()); span.IsRecording() {
					span.SetAttributes(
						attribute.Int("upstreamStatusCode", upstreamStatusCode),
						attribute.Int("resolvedStatusCode", resolvedStatusCode))
				}
				err := eventLogger.LogEvent(
					r.Context(),
					events.Event{
						Name:       codygateway.EventNameCompletionsFinished,
						Source:     act.Source.Name(),
						Identifier: act.ID,
						Metadata: mergeMaps(requestMetadata, map[string]any{
							codygateway.CompletionsEventFeatureMetadataField: feature,
							"model":    gatewayModel,
							"provider": upstreamName,

							// Request details
							"upstream_request_duration_ms": time.Since(upstreamStarted).Milliseconds(),
							"upstream_status_code":         upstreamStatusCode,
							"resolved_status_code":         resolvedStatusCode,

							// Usage details
							"prompt_character_count":     promptCharacterCount,
							"completion_character_count": completionCharacterCount,
						}),
					},
				)
				if err != nil {
					logger.Error("failed to log event", log.Error(err))
				}
			}()

			resp, err := httpClient.Do(req)
			if err != nil {
				// Ignore reporting errors where client disconnected
				if req.Context().Err() == context.Canceled && errors.Is(err, context.Canceled) {
					oteltrace.SpanFromContext(req.Context()).RecordError(err)
					logger.Info("request canceled", log.Error(err))
					return
				}

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

			// Record upstream's status code and decide what we want to send to
			// the client. By default, we just send upstream's status code.
			upstreamStatusCode = resp.StatusCode
			resolvedStatusCode = upstreamStatusCode
			if upstreamStatusCode == http.StatusTooManyRequests {
				// Rewrite 429 to 503 because we share a quota when talking to upstream,
				// and a 429 from upstream should NOT indicate to the client that they
				// should retry. To ensure we are notified when this happens, log this
				// as an error and record the headers that are provided to us.
				var headers bytes.Buffer
				_ = resp.Header.Write(&headers)
				logger.Error("upstream returned 429, rewriting to 503",
					log.Error(errors.New(resp.Status)), // real error needed for Sentry reporting
					log.String("resp.headers", headers.String()))
				resolvedStatusCode = http.StatusServiceUnavailable
			}

			// Write the resolved status code.
			w.WriteHeader(resolvedStatusCode)

			// Set up a buffer to capture the response as it's streamed and sent to the client.
			var responseBuf bytes.Buffer
			respBody := io.TeeReader(resp.Body, &responseBuf)
			// Forward response to client.
			_, _ = io.Copy(w, respBody)

			if upstreamStatusCode >= 200 && upstreamStatusCode < 300 {
				// Pass reader to response transformer to capture token counts.
				completionCharacterCount = parseResponse(body, &responseBuf)

			} else if upstreamStatusCode >= 500 {
				logger.Error("error from upstream",
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

func mergeMaps(dst, src map[string]any) map[string]any {
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
