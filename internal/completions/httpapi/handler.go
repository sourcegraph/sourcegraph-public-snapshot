package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/cody"
	"github.com/sourcegraph/sourcegraph/internal/completions/client"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/internal/telemetry"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// maxRequestDuration is the maximum amount of time a request can take before
// being cancelled.
const maxRequestDuration = time.Minute

func newCompletionsHandler(
	logger log.Logger,
	events *telemetry.EventRecorder,
	feature types.CompletionsFeature,
	rl RateLimiter,
	traceFamily string,
	getModel func(types.CodyCompletionRequestParameters, *conftypes.CompletionsConfig) (string, error),
) http.Handler {
	responseHandler := newSwitchingResponseHandler(logger, feature)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, fmt.Sprintf("unsupported method %s", r.Method), http.StatusMethodNotAllowed)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), maxRequestDuration)
		defer cancel()

		if isEnabled := cody.IsCodyEnabled(ctx); !isEnabled {
			http.Error(w, "cody experimental feature flag is not enabled for current user", http.StatusUnauthorized)
			return
		}

		completionsConfig := conf.GetCompletionsConfig(conf.Get().SiteConfig())
		if completionsConfig == nil {
			http.Error(w, "completions are not configured or disabled", http.StatusInternalServerError)
		}

		var requestParams types.CodyCompletionRequestParameters
		if err := json.NewDecoder(r.Body).Decode(&requestParams); err != nil {
			http.Error(w, "could not decode request body", http.StatusBadRequest)
			return
		}

		// TODO: Model is not configurable but technically allowed in the request body right now.
		var err error
		requestParams.Model, err = getModel(requestParams, completionsConfig)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx, done := Trace(ctx, traceFamily, requestParams.Model, requestParams.MaxTokensToSample).
			WithErrorP(&err).
			WithRequest(r).
			Build()
		defer done()

		completionClient, err := client.Get(
			logger,
			events,
			completionsConfig.Endpoint,
			completionsConfig.Provider,
			completionsConfig.AccessToken,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Check rate limit.
		err = rl.TryAcquire(ctx)
		if err != nil {
			if unwrap, ok := err.(RateLimitExceededError); ok {
				respondRateLimited(w, unwrap)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		responseHandler(ctx, requestParams.CompletionRequestParameters, completionClient, w)
	})
}

func respondRateLimited(w http.ResponseWriter, err RateLimitExceededError) {
	// Rate limit exceeded, write well known headers and return correct status code.
	w.Header().Set("x-ratelimit-limit", strconv.Itoa(err.Limit))
	w.Header().Set("x-ratelimit-remaining", strconv.Itoa(max(err.Limit-err.Used, 0)))
	w.Header().Set("retry-after", err.RetryAfter.Format(time.RFC1123))
	http.Error(w, err.Error(), http.StatusTooManyRequests)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// newSwitchingResponseHandler handles requests to an LLM provider, and wraps the correct
// handler based on the requestParams.Stream flag.
func newSwitchingResponseHandler(logger log.Logger, feature types.CompletionsFeature) func(ctx context.Context, requestParams types.CompletionRequestParameters, cc types.CompletionsClient, w http.ResponseWriter) {
	nonStreamer := newNonStreamingResponseHandler(logger, feature)
	streamer := newStreamingResponseHandler(logger, feature)
	return func(ctx context.Context, requestParams types.CompletionRequestParameters, cc types.CompletionsClient, w http.ResponseWriter) {
		if requestParams.IsStream(feature) {
			streamer(ctx, requestParams, cc, w)
		} else {
			nonStreamer(ctx, requestParams, cc, w)
		}
	}
}

// newStreamingResponseHandler handles streaming requests to an LLM provider,
// It writes events to an SSE stream as they come in.
func newStreamingResponseHandler(logger log.Logger, feature types.CompletionsFeature) func(ctx context.Context, requestParams types.CompletionRequestParameters, cc types.CompletionsClient, w http.ResponseWriter) {
	return func(ctx context.Context, requestParams types.CompletionRequestParameters, cc types.CompletionsClient, w http.ResponseWriter) {
		eventWriter, err := streamhttp.NewWriter(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Always send a final done event so clients know the stream is shutting down.
		defer func() {
			_ = eventWriter.Event("done", map[string]any{})
		}()

		err = cc.Stream(ctx, feature, requestParams,
			func(event types.CompletionResponse) error {
				return eventWriter.Event("completion", event)
			})
		if err != nil {
			l := trace.Logger(ctx, logger)

			logFields := []log.Field{log.Error(err)}
			if errNotOK, ok := types.IsErrStatusNotOK(err); ok {
				if tc := errNotOK.SourceTraceContext; tc != nil {
					logFields = append(logFields,
						log.String("sourceTraceContext.traceID", tc.TraceID),
						log.String("sourceTraceContext.spanID", tc.SpanID))
				}
			}
			l.Error("error while streaming completions", logFields...)

			// Note that we do NOT attempt to forward the status code to the
			// client here, since we are using streamhttp.Writer - see
			// streamhttp.NewWriter for more details. Instead, we send an error
			// event, which clients should check as appropriate.
			if err := eventWriter.Event("error", map[string]string{"error": err.Error()}); err != nil {
				l.Error("error reporting streaming completion error", log.Error(err))
			}
			return
		}
	}
}

// newNonStreamingResponseHandler handles non-streaming requests to an LLM provider,
// awaiting the complete response before writing it back in a structured JSON response
// to the client.
func newNonStreamingResponseHandler(logger log.Logger, feature types.CompletionsFeature) func(ctx context.Context, requestParams types.CompletionRequestParameters, cc types.CompletionsClient, w http.ResponseWriter) {
	return func(ctx context.Context, requestParams types.CompletionRequestParameters, cc types.CompletionsClient, w http.ResponseWriter) {
		completion, err := cc.Complete(ctx, feature, requestParams)
		if err != nil {
			logFields := []log.Field{log.Error(err)}

			// Propagate the upstream headers to the client if available.
			if errNotOK, ok := types.IsErrStatusNotOK(err); ok {
				errNotOK.WriteHeader(w)
				if tc := errNotOK.SourceTraceContext; tc != nil {
					logFields = append(logFields,
						log.String("sourceTraceContext.traceID", tc.TraceID),
						log.String("sourceTraceContext.spanID", tc.SpanID))
				}
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			_, _ = w.Write([]byte(err.Error()))

			trace.Logger(ctx, logger).Error("error on completion", logFields...)
			return
		}

		completionBytes, err := json.Marshal(completion)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(completionBytes)
	}
}
