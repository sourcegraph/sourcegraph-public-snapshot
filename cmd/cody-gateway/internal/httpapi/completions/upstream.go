package completions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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
	"github.com/sourcegraph/sourcegraph/internal/requestclient"
	sgtrace "github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type usageStats struct {
	// characters is the number of characters in the input or response.
	characters int
	// tokens is the number of tokens consumed in the input or response.
	tokens int
}

// upstreamHandlerMethods declares a set of methods that are used throughout the
// lifecycle of a request to an upstream API. All methods are required, and called
// in the order they are defined here.
//
// Methods do not need to be concurrency-safe, as they are only called sequentially.
type upstreamHandlerMethods[ReqT UpstreamRequest] struct {
	// validateRequest can be used to validate the HTTP request before it is sent upstream.
	// Returning a non-nil error will stop further processing and return the given error
	// code, or a 400.
	// Second return value is a boolean indicating whether the request was flagged during validation.
	//
	// The provided logger already contains actor context.
	validateRequest func(context.Context, log.Logger, codygateway.Feature, ReqT) (int, *flaggingResult, error)
	// transformBody can be used to modify the request body before it is sent
	// upstream. To manipulate the HTTP request, use transformRequest.
	//
	// If the upstream supports it, the given identifier string should be
	// provided to assist in abuse detection.
	transformBody func(_ *ReqT, identifier string)
	// transformRequest can be used to modify the HTTP request before it is sent
	// upstream. To manipulate the body, use transformBody.
	transformRequest func(*http.Request)
	// getRequestMetadata should extract details about the request we are sending
	// upstream for validation and tracking purposes. Usage data does not need
	// to be reported here - instead, use parseResponseAndUsage to extract usage,
	// which for some providers we can only know after the fact based on what
	// upstream tells us.
	getRequestMetadata func(context.Context, log.Logger, *actor.Actor, codygateway.Feature, ReqT) (model string, additionalMetadata map[string]any)
	// parseResponseAndUsage should extract details from the response we get back from
	// upstream as well as overall usage for tracking purposes.
	//
	// If data is unavailable, implementations should set relevant usage fields
	// to -1 as a sentinel value.
	parseResponseAndUsage func(log.Logger, ReqT, io.Reader) (promptUsage, completionUsage usageStats)
}

type UpstreamRequest interface {
	GetModel() string
	ShouldStream() bool
}

func makeUpstreamHandler[ReqT UpstreamRequest](
	baseLogger log.Logger,
	eventLogger events.Logger,
	rs limiter.RedisStore,
	rateLimitNotifier notify.RateLimitNotifier,
	httpClient httpcli.Doer,

	// upstreamName is the name of the upstream provider. It MUST match the
	// provider names defined clientside, i.e. "anthropic" or "openai".
	upstreamName string,

	upstreamAPIURL func(feature codygateway.Feature) string,
	allowedModels []string,

	methods upstreamHandlerMethods[ReqT],

	// defaultRetryAfterSeconds sets the retry-after policy on upstream rate
	// limit events in case a retry-after is not provided by the upstream
	// response.
	defaultRetryAfterSeconds int,
	autoFlushStreamingResponses bool,
) http.Handler {
	baseLogger = baseLogger.Scoped(upstreamName).
		// This URL is used only for logging reason so we default to the chat endpoint
		With(log.String("upstream.url", upstreamAPIURL(codygateway.FeatureChatCompletions)))

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

			// TODO: Investigate using actor propagation handler for extracting
			// this. We had some issues before getting that to work, so for now
			// just stick with what we've seen working so far.
			sgActorID := r.Header.Get("X-Sourcegraph-Actor-UID")
			sgActorAnonymousUID := r.Header.Get("X-Sourcegraph-Actor-Anonymous-UID")

			// Build logger for lifecycle of this request with lots of details.
			logger := act.Logger(sgtrace.Logger(r.Context(), baseLogger)).With(
				append(
					requestclient.FromContext(r.Context()).LogFields(),
					// Sourcegraph actor details
					log.String("sg.actorID", sgActorID),
					log.String("sg.anonymousID", sgActorAnonymousUID),
				)...,
			)

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
			status, flaggingResult, err := methods.validateRequest(r.Context(), logger, feature, body)
			if err != nil {
				if status == 0 {
					response.JSONError(logger, w, http.StatusBadRequest, errors.Wrap(err, "invalid request"))
				}
				if flaggingResult.IsFlagged() && flaggingResult.shouldBlock {
					requestMetadata := getFlaggingMetadata(flaggingResult, act)
					err := eventLogger.LogEvent(
						r.Context(),
						events.Event{
							Name:       codygateway.EventNameRequestBlocked,
							Source:     act.Source.Name(),
							Identifier: act.ID,
							Metadata: mergeMaps(requestMetadata, map[string]any{
								codygateway.CompletionsEventFeatureMetadataField: feature,
								"model":    fmt.Sprintf("%s/%s", upstreamName, body.GetModel()),
								"provider": upstreamName,

								// Response details
								"resolved_status_code": status,

								// Request metadata
								"prompt_token_count":   flaggingResult.promptTokenCount,
								"max_tokens_to_sample": flaggingResult.maxTokensToSample,

								// Actor details, specific to the actor Source
								"sg_actor_id":            sgActorID,
								"sg_actor_anonymous_uid": sgActorAnonymousUID,
							}),
						},
					)
					if err != nil {
						logger.Error("failed to log event", log.Error(err))
					}
				}

				response.JSONError(logger, w, status, err)
				return
			}

			// identifier that can be provided to upstream for abuse detection
			// has the format '$ACTOR_ID:$SG_ACTOR_ID'. The latter is anonymized
			// (specific per-instance)
			identifier := fmt.Sprintf("%s:%s", act.ID, sgActorID)
			methods.transformBody(&body, identifier)

			// Re-marshal the payload for upstream to unset metadata and remove any properties
			// not known to us.
			upstreamPayload, err := json.Marshal(body)
			if err != nil {
				response.JSONError(logger, w, http.StatusInternalServerError, errors.Wrap(err, "failed to marshal request body"))
				return
			}

			// Create a new request to send upstream, making sure we retain the same context.
			req, err := http.NewRequestWithContext(r.Context(), http.MethodPost,  upstreamAPIURL(feature), bytes.NewReader(upstreamPayload))
			if err != nil {
				response.JSONError(logger, w, http.StatusInternalServerError, errors.Wrap(err, "failed to create request"))
				return
			}

			// Run the request transformer.
			methods.transformRequest(req)

			// Retrieve metadata from the initial request.
			model, requestMetadata := methods.getRequestMetadata(r.Context(), logger, act, feature, body)

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
				// promptUsage and completionUsage are extracted from parseResponseAndUsage.
				promptUsage, completionUsage usageStats
			)
			defer func() {
				if span := oteltrace.SpanFromContext(r.Context()); span.IsRecording() {
					span.SetAttributes(
						attribute.Int("upstreamStatusCode", upstreamStatusCode),
						attribute.Int("resolvedStatusCode", resolvedStatusCode))
				}
				if flaggingResult.IsFlagged() {
					requestMetadata = mergeMaps(requestMetadata, getFlaggingMetadata(flaggingResult, act))
				}
				usageData := map[string]any{
					"prompt_character_count":     promptUsage.characters,
					"prompt_token_count":         promptUsage.tokens,
					"completion_character_count": completionUsage.characters,
					"completion_token_count":     completionUsage.tokens,
				}
				for k, v := range usageData {
					// Drop usage fields that are invalid/unimplemented. All
					// usageData fields are ints - we use map[string]any for
					// convenience with mergeMaps utility.
					if n, _ := v.(int); n < 0 {
						delete(usageData, k)
					}
				}
				err := eventLogger.LogEvent(
					r.Context(),
					events.Event{
						Name:       codygateway.EventNameCompletionsFinished,
						Source:     act.Source.Name(),
						Identifier: act.ID,
						Metadata: mergeMaps(requestMetadata, usageData, map[string]any{
							codygateway.CompletionsEventFeatureMetadataField: feature,
							"model":    gatewayModel,
							"provider": upstreamName,

							// Request details
							"upstream_request_duration_ms": time.Since(upstreamStarted).Milliseconds(),
							"upstream_status_code":         upstreamStatusCode,
							"resolved_status_code":         resolvedStatusCode,

							// Actor details, specific to the actor Source
							"sg_actor_id":            sgActorID,
							"sg_actor_anonymous_uid": sgActorAnonymousUID,
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
					oteltrace.SpanFromContext(req.Context()).
						SetStatus(codes.Error, err.Error())
					logger.Info("request canceled", log.Error(err))
					return
				}

				// More user-friendly message for timeouts
				if errors.Is(err, context.DeadlineExceeded) {
					resolvedStatusCode = http.StatusGatewayTimeout
					response.JSONError(logger, w, resolvedStatusCode,
						errors.Newf("request to upstream provider %s timed out", upstreamName))
					return
				}

				resolvedStatusCode = http.StatusInternalServerError
				response.JSONError(logger, w, resolvedStatusCode,
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
				// should liberally retry until the rate limit is lifted. To ensure we are
				// notified when this happens, log this as an error and record the headers
				// that are provided to us.
				var headers bytes.Buffer
				_ = resp.Header.Write(&headers)
				logger.Error("upstream returned 429, rewriting to 503",
					log.Error(errors.New(resp.Status)), // real error needed for Sentry reporting
					log.String("resp.headers", headers.String()))
				resolvedStatusCode = http.StatusServiceUnavailable
			}

			// This handles upstream 429 responses as well, since they get
			// resolved to http.StatusServiceUnavailable.
			if resolvedStatusCode == http.StatusServiceUnavailable {
				// Propagate retry-after in case it is handle-able by the client,
				// or write our default. 503 errors can have retry-after as well.
				if upstreamRetryAfter := resp.Header.Get("retry-after"); upstreamRetryAfter != "" {
					w.Header().Set("retry-after", upstreamRetryAfter)
				} else {
					w.Header().Set("retry-after", strconv.Itoa(defaultRetryAfterSeconds))
				}
			}

			// Write the resolved status code.
			w.WriteHeader(resolvedStatusCode)

			// Set up a buffer to capture the response as it's streamed and sent to the client.
			var responseBuf bytes.Buffer
			respBody := io.TeeReader(resp.Body, &responseBuf)
			// if this is a streaming request, we want to flush ourselves instead of leaving that to the http.Server
			// (so events are sent to the client as soon as possible)
			var responseWriter io.Writer = w
			if autoFlushStreamingResponses && body.ShouldStream() && feature == codygateway.FeatureCodeCompletions {
				if fw, err := response.NewAutoFlushingWriter(w); err == nil {
					responseWriter = fw
				} else {
					// We can't stream the response, but it's better to write it without streaming that fail, so we just log the error
					logger.Error("failed to create auto-flushing writer", log.Error(err))
				}
			}
			_, _ = io.Copy(responseWriter, respBody)

			if upstreamStatusCode >= 200 && upstreamStatusCode < 300 {
				// Pass reader to response transformer to capture token counts.
				promptUsage, completionUsage = methods.parseResponseAndUsage(logger, body, &responseBuf)
			} else if upstreamStatusCode >= 500 {
				logger.Error("error from upstream",
					log.Int("status_code", upstreamStatusCode))
			}
		}))
}

func getFlaggingMetadata(flaggingResult *flaggingResult, act *actor.Actor) map[string]any {
	requestMetadata := map[string]any{}

	requestMetadata["flagged"] = true
	flaggingMetadata := map[string]any{
		"reason":       flaggingResult.reasons,
		"should_block": flaggingResult.shouldBlock,
	}

	if act.IsDotComActor() {
		flaggingMetadata["prompt_prefix"] = flaggingResult.promptPrefix
	}
	requestMetadata["flagging_result"] = flaggingMetadata
	return requestMetadata
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

func mergeMaps(dst map[string]any, srcs ...map[string]any) map[string]any {
	for _, src := range srcs {
		for k, v := range src {
			dst[k] = v
		}
	}
	return dst
}

type flaggingResult struct {
	shouldBlock       bool
	reasons           []string
	promptPrefix      string
	maxTokensToSample int
	promptTokenCount  int
}

func (f *flaggingResult) IsFlagged() bool {
	return f != nil
}
