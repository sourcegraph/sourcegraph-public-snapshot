package services

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
	oteltrace "go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/slices"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/notify"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/response"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	sgtrace "github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Note: modeled very closely off of completions/upstream.go
type bodyTransformer[T any] func(*T)
type requestTransformer func(*http.Request)
type requestMetadataRetriever[T any] func(T) (metadata map[string]any)
type responseParser[T any] func(T, io.Reader) (moreMetadata map[string]any)

// upstreamHandlerMethods declares a set of methods that are used throughout the
// lifecycle of a request to an upstream API. All methods are required.
type upstreamHandlerMethods[ReqT any] struct {
	// transformBody can be used to modify the request body before it is sent
	// upstream. To manipulate the HTTP request, use transformRequest.
	transformBody bodyTransformer[ReqT]
	// transformRequest can be used to modify the HTTP request before it is sent
	// upstream. To manipulate the body, use transformBody.
	transformRequest requestTransformer
	//  should extract details about the request we are sending
	// upstream for validation and tracking purposes.
	requestMetadataRetriever[ReqT]
	// parseResponse should extract details from the response we get back from
	// upstream for tracking purposes.
	parseResponse responseParser[ReqT]
}

func makeUpstreamHandler[ReqT any](
	baseLogger log.Logger,
	eventLogger events.Logger,
	rs limiter.RedisStore,
	rateLimitNotifier notify.RateLimitNotifier,
	serviceProvider string,
	serviceName string,
	upstreamAPIURL string,
	allowedServices []string,
	methods upstreamHandlerMethods[ReqT],
) http.Handler {
	baseLogger = baseLogger.Scoped(serviceProvider, fmt.Sprintf("%s upstream handler", serviceProvider)).
		With(log.String("upstream.url", upstreamAPIURL))

	return rateLimit(
		baseLogger,
		eventLogger,
		limiter.NewPrefixRedisStore("rate_limit:", rs),
		rateLimitNotifier,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			act := actor.FromContext(r.Context())
			logger := act.Logger(sgtrace.Logger(r.Context(), baseLogger))

			gatewayFeatureService := formatServiceFeature(serviceProvider, serviceName)

			// This will never be nil as the rate limiter middleware checks this before.
			// TODO: Should we read the rate limit from context, and store it in the rate
			// limiter to make this less dependent on these two logics to remain the same?
			rateLimit, ok := act.RateLimits[codygateway.Services]
			if !ok {
				response.JSONError(logger, w, http.StatusInternalServerError, errors.Newf("rate limit for %q not found", string(feature)))
				return
			}

			// Parse the request body.
			var body ReqT
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				response.JSONError(logger, w, http.StatusBadRequest, errors.Wrap(err, "failed to parse request body"))
				return
			}

			methods.transformBody(&body)

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
			methods.transformRequest(req)

			// Retrieve metadata from the initial request.
			requestMetadata := methods.(body)

			// Check if the specific service feature is allowed - note we use a data model designed for AI-based services
			// So the field is called "AllowedModels"
			if allowed := intersection(allowedServices, rateLimit.AllowedModels); !isAllowedService(allowed, gatewayFeatureService) {
				response.JSONError(logger, w, http.StatusBadRequest,
					errors.Newf("service feature %q is not allowed, allowed: [%s]",
					gatewayFeatureService, strings.Join(allowed, ", ")))
				return
			}

			var (
				upstreamStarted        = time.Now()
				upstreamStatusCode int = -1
				// resolvedStatusCode is the status code that we returned to the
				// client - in most case it is the same as upstreamStatusCode,
				// but sometimes we write something different.
				resolvedStatusCode int = -1
				responseMetadata   map[string]any
			)
			defer func() {
				err := eventLogger.LogEvent(
					r.Context(),
					events.Event{
						Name:       codygateway.EventNameServiceFinished,
						Source:     act.Source.Name(),
						Identifier: act.ID,
						Metadata: mergeMaps(requestMetadata, responseMetadata, map[string]any{
							codygateway.CompletionsEventFeatureMetadataField:  codygateway.Services,
							"provider": serviceProvider,
							"service":  serviceName,

							// Request details
							"upstream_request_duration_ms": time.Since(upstreamStarted).Milliseconds(),
							"upstream_status_code":         upstreamStatusCode,
							"resolved_status_code":         resolvedStatusCode,
						}),
					},
				)
				if err != nil {
					logger.Error("failed to log event", log.Error(err))
				}
			}()

			resp, err := httpcli.ExternalDoer.Do(req)
			if err != nil {
				// Ignore reporting errors where client disconnected
				if req.Context().Err() == context.Canceled && errors.Is(err, context.Canceled) {
					oteltrace.SpanFromContext(req.Context()).RecordError(err)
					logger.Info("request canceled", log.Error(err))
					return
				}

				response.JSONError(logger, w, http.StatusInternalServerError,
					errors.Wrapf(err, "failed to make request to upstream provider %s", serviceProvider))
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
				// Pass reader to response transformer to capture metadata
				responseMetadata = methods.parseResponse(body, &responseBuf)

			} else if upstreamStatusCode >= 500 {
				logger.Error("error from upstream",
					log.Int("status_code", upstreamStatusCode))
			}
		}))
}

func isAllowedService(allowedServices []string, model string) bool {
	for _, m := range allowedServices {
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

// Format as "SERVICE PROVIDER/FEATURE"
// This is the format that the gateway expects for allow list
func formatServiceFeature(serviceProvider: string, featureName: string): string {
 return fmt.Sprintf("%s/%s", serviceProvider, featureName)
}
