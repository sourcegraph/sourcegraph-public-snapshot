package completions

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Masterminds/semver"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/guardrails"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/accesstoken"
	sgactor "github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/hashutil"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/cody"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/modelconfig"
	"github.com/sourcegraph/sourcegraph/internal/completions/client"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/internal/telemetry"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// maxRequestDuration is the maximum amount of time a request can take before
// being cancelled as DeadlineExceeded.
const maxRequestDuration = 1 * time.Minute

var timeToFirstEventMetrics = metrics.NewREDMetrics(
	prometheus.DefaultRegisterer,
	"completions_stream_first_event",
	metrics.WithLabels("model"),
	metrics.WithDurationBuckets([]float64{0.25, 0.5, 0.75, 1.0, 2.0, 3.0, 4.0, 6.0, 8.0, 10.0, 12.0, 14.0, 16.0, 18.0, 20.0, 25.0, 30.0}),
)

func newCompletionsHandler(
	logger log.Logger,
	db database.DB,
	userStore database.UserStore,
	accessTokenStore database.AccessTokenStore,
	events *telemetry.EventRecorder,
	grAttributionTest guardrails.AttributionTest,
	feature types.CompletionsFeature,
	rl RateLimiter,
	traceFamily string,
	getModel getModelFn,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		// Set the context timeout: use the timeout from the request header if provided,
		// otherwise use the default maximum request duration.
		ctxTimeout := maxRequestDuration
		if v := r.Header.Get("X-Timeout-Ms"); v != "" {
			if t, err := strconv.Atoi(v); err != nil {
				logger.Warn("error parsing X-Timeout-Ms header", log.Error(err))
			} else {
				ctxTimeout = time.Duration(t) * time.Millisecond
			}
		}
		ctx, cancel := context.WithTimeout(r.Context(), ctxTimeout)
		defer cancel()

		// First check that Cody is enabled for this Sourcegraph instance.
		if isEnabled, reason := cody.IsCodyEnabled(ctx, db); !isEnabled {
			errResponse := fmt.Sprintf("cody is not enabled: %s", reason)
			http.Error(w, errResponse, http.StatusUnauthorized)
			return
		}

		var version types.CompletionsVersion
		switch versionParam := r.URL.Query().Get("api-version"); versionParam {
		case "":
			version = types.CompletionsVersionLegacy
		case "1":
			version = types.CompletionsV1
		default:
			logger.Warn(
				"blocking request because unrecognized CompletionsVersion API param",
				log.String("version", versionParam))
			http.Error(w, "Unsupported API Version (Please update your client)", http.StatusNotAcceptable)
			return
		}

		// Enterprise customers may define instance-wide Cody context filters (aka Cody Ignore) in the site config.
		// To ensure Cody clients respect these restrictions, we enforce the minimum supported client version.
		isDotcom := dotcom.SourcegraphDotComMode()
		if !isDotcom {
			if err := checkClientCodyIgnoreCompatibility(ctx, db, r); err != nil {
				logger.Info("rejecting request due to CodyIngore compat", log.Error(err))
				http.Error(w, err.Error(), err.statusCode)
				return
			}
		}

		// We don't perform any sort of validation. So we would silently accept a totally bogus
		// JSON payload. And just have a zero-value CompletionRequestParameters, e.g. no prompt.
		var requestParams types.CodyCompletionRequestParameters
		if err := json.NewDecoder(r.Body).Decode(&requestParams); err != nil {
			logger.Warn("malformed CodyCompletionRequestParameters", log.Error(err))
			http.Error(w, "could not decode request body", http.StatusBadRequest)
			return
		}

		// Load the current LLM model configuration for the Sourcegraph instance.
		modelConfigSvc := modelconfig.Get()
		currentModelConfig, err := modelConfigSvc.Get()
		if err != nil {
			logger.Error("fetching current LLM model configuration", log.Error(err))
			http.Error(w, "internal error loading model configuration.", http.StatusInternalServerError)
			return
		}

		// Load the Provider and Model configuration data. This is surprisingly tricky, because of
		// various contextual defaults and/or checking the user has access to the model, etc.
		providerConfig, modelConfig, err := resolveRequestedModel(ctx, logger, currentModelConfig, requestParams, getModel)
		if err != nil {
			// NOTE: We return the raw error to the user assuming that it contains relevant
			// user-facing diagnostic information, and doesn't leak any internal details.
			logger.Info("error resolving model", log.Error(err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx, done := Trace(ctx, traceFamily, modelConfig.ModelName, requestParams.MaxTokensToSample).
			WithErrorP(&err).
			WithRequest(r).
			Build()
		defer done()

		// Will the current LLM request be sent to Cody Gateway?
		var willBeSentToCodyGateway bool
		if ssConfig := providerConfig.ServerSideConfig; ssConfig != nil {
			willBeSentToCodyGateway = ssConfig.SourcegraphProvider != nil
		}

		modelConfigInfo := types.ModelConfigInfo{
			Provider: *providerConfig,
			Model:    *modelConfig,

			// For Cody Enterprise, the LLM request will be sent using the credentials
			// found in the site config. However, for Cody Pro, we use an access token
			// tied to the calling user.
			CodyProUserAccessToken: nil,
		}
		if isDotcom {
			// Confirm the required provider is Cody Gateway. We may later support LLM models for Cody Pro
			// that are not proxied through Cody Gateway. But for now, it is the only case and we are just
			// confirming things are configured the way we expect.
			if !willBeSentToCodyGateway {
				logger.Error("the configuration is wrong, dotcom received request that will NOT be routed to Cody Gateway")
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			// Get the Cody Gateway credentials to send.
			codyProUserAccessToken, err := getCallersCodyProAccessToken(ctx, logger, accessTokenStore, r)
			if err != nil {
				logger.Error("error getting Cody Free/Pro user's access token", log.Error(err))
				http.Error(w, "internal error creating gateway credentials", http.StatusInternalServerError)
				return
			}
			modelConfigInfo.CodyProUserAccessToken = &codyProUserAccessToken
		}

		// Finally! With all the necessary information we can now create the CompletionsClient for
		// contacting the LLM and responding to the request.
		completionClient, err := client.Get(
			logger,
			events,
			modelConfigInfo)
		l := trace.Logger(ctx, logger)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !isDotcom || !willBeSentToCodyGateway {
			// Check rate limit.
			err = rl.TryAcquire(ctx)
			if err != nil {
				if unwrap, ok := err.(RateLimitExceededError); ok {
					actor := sgactor.FromContext(ctx)
					user, err := actor.User(ctx, userStore)
					if err != nil {
						l.Error("Error while fetching user", log.Error(err))
						http.Error(w, "Internal server error", http.StatusInternalServerError)
						return
					}
					subscription, err := cody.SubscriptionForUser(ctx, db, *user)
					if err != nil {
						l.Error("Error while fetching user's cody subscription", log.Error(err))
						http.Error(w, "Internal server error", http.StatusInternalServerError)
						return
					}

					respondRateLimited(w, unwrap, isDotcom, subscription.ApplyProRateLimits)
					return
				}
				l.Warn("Rate limit error", log.Error(err))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		}

		// Finally serve the request.
		fullCompletionRequest := types.CompletionRequest{
			Feature:         feature,
			ModelConfigInfo: modelConfigInfo,
			Parameters:      requestParams.CompletionRequestParameters,
			Version:         version,
		}
		if requestParams.IsStream(feature) {
			serveStreamingResponse(
				w,
				ctx, db, logger,
				fullCompletionRequest, completionClient, grAttributionTest)
		} else {
			serveSyncResponse(
				w,
				ctx, db, logger,
				fullCompletionRequest, completionClient, grAttributionTest)
		}
	})
}

// getCallersCodyProAccessToken pulls out the credential information for the calling user, and
// will return an access token suitable for passing to Cody Gateway.
func getCallersCodyProAccessToken(
	ctx context.Context, logger log.Logger, accessTokenStore database.AccessTokenStore, r *http.Request) (string, error) {

	callerAPIToken, _, err := authz.ParseAuthorizationHeader(r.Header.Get("Authorization"))
	if err == nil {
		gatewayAccessToken, err := accesstoken.GenerateDotcomUserGatewayAccessToken(callerAPIToken)
		if err != nil {
			trace.Logger(ctx, logger).Info("Access token generation failed", log.Error(err))
			return "", errors.New("access token generation failed")
		}
		return gatewayAccessToken, nil
	}

	// If there was an error fetching the dotcom user's token from the auth header,
	// try to just create one some other way.
	if r.Header.Get("Authorization") != "" || sgactor.FromContext(ctx) == nil {
		trace.Logger(ctx, logger).Info("Error parsing auth header", log.Error(err))
		return "", errors.New("error parsing auth header")
	}

	// Get or create an internal token to use, scoped to the calling actor.
	actor := sgactor.FromContext(ctx)
	apiTokenSha256, err := accessTokenStore.GetOrCreateInternalToken(ctx, actor.UID, []string{"user:all"})
	if err != nil {
		trace.Logger(ctx, logger).Info("Error creating internal access token", log.Error(err))
		return "", errors.New("creating access token")
	}

	// Convert the user's sha256-encoded access token to an "sgd_" token for Cody Gateway.
	// Note: we can't use accesstoken.GenerateDotcomUserGatewayAccessToken here because
	// we only need to hash this once, not twice, as this is already an SHA256-encoding
	// of the original token.
	newAccessToken := accesstoken.DotcomUserGatewayAccessTokenPrefix + hex.EncodeToString(hashutil.ToSHA256Bytes(apiTokenSha256))
	return newAccessToken, nil
}

func respondRateLimited(w http.ResponseWriter, err RateLimitExceededError, isDotcom, isProUser bool) {
	// Rate limit exceeded, write well known headers and return correct status code.
	w.Header().Set("x-ratelimit-limit", strconv.Itoa(err.Limit))
	w.Header().Set("x-ratelimit-remaining", strconv.Itoa(max(err.Limit-err.Used, 0)))
	w.Header().Set("retry-after", err.RetryAfter.Format(time.RFC1123))
	if isDotcom {
		if isProUser {
			w.Header().Set("x-is-cody-pro-user", "true")
		} else {
			w.Header().Set("x-is-cody-pro-user", "false")
		}
	}
	http.Error(w, err.Error(), http.StatusTooManyRequests)
}

// newStreamingResponseHandler handles streaming requests to an LLM provider,
// It writes events to an SSE stream as they come in.
func serveStreamingResponse(
	w http.ResponseWriter,
	ctx context.Context, db database.DB, logger log.Logger,
	compRequest types.CompletionRequest, cc types.CompletionsClient, test guardrails.AttributionTest) {
	var eventWriter = sync.OnceValue[*streamhttp.Writer](func() *streamhttp.Writer {
		// NOTE: The HTTP response defaults to 200 if not set explicitly before writing the response.
		// This means that this function will never serve a non-OK response. (Which makes sense, since
		// we don't know ahead of time if some later SSE event will fail.
		eventWriter, err := streamhttp.NewWriter(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil
		}
		return eventWriter
	})

	// Isolate writing events.
	var mu sync.Mutex
	writeEvent := func(name string, data any) (err error) {
		// Attribution search is currently panicing. The main hypothesis
		// is calling writeEvent on a finished request in the case of an
		// error. Rather than panicing the whole process we convert it
		// into an error which will get logged.
		//
		// https://github.com/sourcegraph/sourcegraph/issues/60439
		defer func() {
			if rec := recover(); rec != nil {
				err = errors.WithStack(errors.Errorf("recovered panic in completions writeEvent: %v", rec))
			}
		}()

		mu.Lock()
		defer mu.Unlock()
		if ev := eventWriter(); ev != nil {
			return ev.Event(name, data)
		}
		return nil
	}

	// Always send a final done event so clients know the stream is shutting down.
	firstEventObserved := false
	defer func() {
		if firstEventObserved {
			_ = writeEvent("done", map[string]any{})
		}
	}()
	start := time.Now()
	eventSink := func(e types.CompletionResponse) error {
		return writeEvent("completion", e)
	}
	attributionErrorLog := func(err error) {
		l := trace.Logger(ctx, logger)
		if err := writeEvent("attribution-error", map[string]string{"error": err.Error()}); err != nil {
			l.Error("error reporting attribution error", log.Error(err))
		} else {
			return
		}
		l.Error("attribution error", log.Error(err))
	}
	f := guardrails.NoopCompletionsFilter(eventSink)
	if cf := conf.GetConfigFeatures(conf.SiteConfig()); cf != nil && cf.Attribution &&
		featureflag.FromContext(ctx).GetBoolOr("autocomplete-attribution", true) {
		factory := guardrails.NewCompletionsFilter
		// TODO(#61828) - Validate & cleanup:
		// 1.  If experiments are successful on S2 and we do not see any panics,
		//     please switch the feature flag default value to true.
		// 2.  Afterwards cleanup the implementation and only use V2 completion filter.
		//     Remove v1 implementation completely.
		if featureflag.FromContext(ctx).GetBoolOr("autocomplete-attribution-v2", false) {
			factory = guardrails.NewCompletionsFilter2
		}
		ff, err := factory(guardrails.CompletionsFilterConfig{
			Sink:             eventSink,
			Test:             test,
			AttributionError: attributionErrorLog,
		})
		if err != nil {
			attributionErrorLog(err)
		} else {
			f = ff
		}
	}

	// Build and send the completions request.
	modelName := compRequest.ModelConfigInfo.Model.ModelName
	sendEventFn := func(event types.CompletionResponse) error {
		if !firstEventObserved {
			firstEventObserved = true
			timeToFirstEventMetrics.Observe(time.Since(start).Seconds(), 1, nil, modelName)
		}
		return f.Send(ctx, event)
	}
	err := cc.Stream(ctx, logger, compRequest, sendEventFn)
	if err != nil {
		l := trace.Logger(ctx, logger)

		logFields := []log.Field{log.Error(err)}
		if errNotOK, ok := types.IsErrStatusNotOK(err); ok {
			if !firstEventObserved && errNotOK.StatusCode == http.StatusTooManyRequests {
				actor := sgactor.FromContext(ctx)
				user, err := actor.User(ctx, db.Users())
				if err != nil {
					l.Error("Error while fetching user", log.Error(err))
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

				subscription, err := cody.SubscriptionForUser(ctx, db, *user)
				if err != nil {
					l.Error("Error while fetching user's cody subscription", log.Error(err))
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

				isDotcom := dotcom.SourcegraphDotComMode()
				if isDotcom {
					if subscription.ApplyProRateLimits {
						w.Header().Set("x-is-cody-pro-user", "true")
					} else {
						w.Header().Set("x-is-cody-pro-user", "false")
					}
				}
				errNotOK.WriteHeader(w)
				return

			}
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
		if !firstEventObserved {
			firstEventObserved = true
			modelName := compRequest.ModelConfigInfo.Model.ModelName
			timeToFirstEventMetrics.Observe(time.Since(start).Seconds(), 1, nil, modelName)
		}
		if err := writeEvent("error", map[string]string{"error": err.Error()}); err != nil {
			l.Error("error reporting streaming completion error", log.Error(err))
		}
		return
	}
	if f != nil { // if autocomplete-attribution enabled
		if err := f.WaitDone(ctx); err != nil {
			l := trace.Logger(ctx, logger)
			if err := writeEvent("error", map[string]string{"error": err.Error()}); err != nil {
				l.Error("error reporting streaming completion error", log.Error(err))
			}
		}
	}
}

// newNonStreamingResponseHandler handles non-streaming requests to an LLM provider,
// awaiting the complete response before writing it back in a structured JSON response
// to the client.
func serveSyncResponse(
	w http.ResponseWriter,
	ctx context.Context, db database.DB, logger log.Logger,
	compRequest types.CompletionRequest, cc types.CompletionsClient, _ guardrails.AttributionTest) {

	completion, err := cc.Complete(ctx, logger, compRequest)
	if err != nil {
		logFields := []log.Field{log.Error(err)}

		// Propagate the upstream headers to the client if available.
		if errNotOK, ok := types.IsErrStatusNotOK(err); ok {
			errNotOK.WriteHeader(w)

			if errNotOK.StatusCode == http.StatusTooManyRequests {
				actor := sgactor.FromContext(ctx)
				user, err := actor.User(ctx, db.Users())
				if err != nil {
					logger.Error("Error while fetching user", log.Error(err))
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
				subscription, err := cody.SubscriptionForUser(ctx, db, *user)
				if err != nil {
					logger.Error("Error while fetching user's cody subscription", log.Error(err))
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

				isDotcom := dotcom.SourcegraphDotComMode()
				if isDotcom {
					if subscription.ApplyProRateLimits {
						w.Header().Set("x-is-cody-pro-user", "true")
					} else {
						w.Header().Set("x-is-cody-pro-user", "false")
					}
				}
				return

			}

			if tc := errNotOK.SourceTraceContext; tc != nil {
				logFields = append(logFields,
					log.String("sourceTraceContext.traceID", tc.TraceID),
					log.String("sourceTraceContext.spanID", tc.SpanID))
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			logger.Error("failed to write", log.Error(err))
		}

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

type codyIgnoreCompatibilityError struct {
	reason     string
	statusCode int
}

func (e *codyIgnoreCompatibilityError) Error() string {
	// prefix value is used to identify specific errors in the Cody clients codebases.
	// When changing its value be sure to update the clients code.
	const prefix = "ClientCodyIgnoreCompatibilityError"
	return fmt.Sprintf("%s: %s", prefix, e.reason)
}

// checkClientCodyIgnoreCompatibility checks if the client version respects Cody context filters (aka Cody Ignore) defined in the site config.
// A non-nil codyIgnoreCompatibilityError implies that the HTTP request should be rejected with the error text and code from the returned error.
// Error text is safe to be surfaced to end user.
func checkClientCodyIgnoreCompatibility(ctx context.Context, db database.DB, r *http.Request) *codyIgnoreCompatibilityError {
	// If Cody context filters are not defined on the instance, we do not restrict client version.
	// Because the site hasn't configured Cody Ignore, no need to enforce it.
	if conf.SiteConfig().CodyContextFilters == nil {
		return nil
	}

	clientName := types.CodyClientName(r.URL.Query().Get("client-name"))
	if clientName == "" {
		return &codyIgnoreCompatibilityError{
			reason:     "\"client-name\" query param is required.",
			statusCode: http.StatusNotAcceptable,
		}
	}

	// If cody-context-filters-clients-test-mode feature flag is enabled, the client version constaint
	// allows pre-release versions (see https://pkg.go.dev/github.com/Masterminds/semver#readme-working-with-pre-release-versions).
	// Intended for development use only.
	flag, err := db.FeatureFlags().GetFeatureFlag(ctx, "cody-context-filters-clients-test-mode")
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return &codyIgnoreCompatibilityError{
			reason:     "Failed to get feature flag value.",
			statusCode: http.StatusInternalServerError,
		}
	}
	isClientsTestMode := flag != nil && flag.Bool.Value

	// clientVersionConstraint defines the minimum client version required to support Cody Ignore.
	type clientVersionConstraint struct {
		client     types.CodyClientName
		constraint string // represents client version constraint following the semver spec
	}
	var cvc clientVersionConstraint
	switch clientName {
	case types.CodyClientVscode:
		cvc = clientVersionConstraint{client: clientName, constraint: ">= 1.20.0"}
		if isClientsTestMode {
			// set the constraint to a lower pre-release version to enable testing
			cvc.constraint = ">= 1.16.0-0"
		}
	case types.CodyClientJetbrains:
		cvc = clientVersionConstraint{client: clientName, constraint: ">= 6.0.0"}
		if isClientsTestMode {
			// set the constraint to a lower pre-release version to enable testing
			cvc.constraint = ">= 5.5.8-0"
		}
	case types.CodyClientWeb:
		// Don't require client version for Web because it's versioned with the Sourcegraph instance.
		return nil
	default:
		// By default, allow requests from any client on any version. We only
		// want to reject requests from older client versions that we know
		// definitely don't support context filters.
		// All agent-based clients (JetBrains, Eclipse, Visual Studio) support
		// context filters out of the box since the original support was added
		// for JetBrains GA in May 2024.
		cvc = clientVersionConstraint{client: clientName, constraint: ">= 0.0.0-0"}
	}

	clientVersion := r.URL.Query().Get("client-version")
	if clientVersion == "" {
		return &codyIgnoreCompatibilityError{
			reason:     "\"client-version\" query param is required.",
			statusCode: http.StatusNotAcceptable,
		}
	}

	c, err := semver.NewConstraint(cvc.constraint)
	if err != nil {
		return &codyIgnoreCompatibilityError{
			reason:     fmt.Sprintf("Cody for %s version constraint %q doesn't follow semver spec.", cvc.client, cvc.constraint),
			statusCode: http.StatusInternalServerError,
		}
	}

	v, err := semver.NewVersion(clientVersion)
	if err != nil {
		return &codyIgnoreCompatibilityError{
			reason:     fmt.Sprintf("Cody for %s version %q doesn't follow semver spec.", cvc.client, clientVersion),
			statusCode: http.StatusBadRequest,
		}
	}

	ok := c.Check(v)
	if !ok {
		return &codyIgnoreCompatibilityError{
			reason:     fmt.Sprintf("Cody for %s version %q doesn't match version constraint %q. Please upgrade your client.", cvc.client, clientVersion, cvc.constraint),
			statusCode: http.StatusNotAcceptable,
		}
	}

	return nil
}
