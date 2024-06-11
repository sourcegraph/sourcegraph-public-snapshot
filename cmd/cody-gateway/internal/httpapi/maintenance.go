package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/httpapi/requestlogger"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/shared/config"
	"github.com/sourcegraph/sourcegraph/internal/instrumentation"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"

	"github.com/sourcegraph/sourcegraph/internal/sams"
)

const samsScopeFlaggedPromptRead = "cody_gateway::flaggedprompts::read"

// NewMaintenanceHandler registers maintenance-related endpoints. These are APIs for
// other services to call in order to inspect or update Cody Gateway internals.
//
// The registered endpoints require a SAMS authentication token, and Cody Gateway scopes.
func NewMaintenanceHandler(
	baseLogger log.Logger, next http.Handler, config *config.Config, redisKV redispool.KeyValue) http.Handler {
	// Do nothing if no SAMS configuration is provided.
	if err := config.SAMSClientConfig.Validate(); err != nil {
		baseLogger.Warn("no SAMS client config provided; not registering maintenance endpoints",
			log.Error(err))
		return next
	}

	logger := baseLogger.Scoped("Maintenance")

	// Create the SAMS Client.
	samsAPIURL := pointers.Deref(
		config.SAMSClientConfig.ConnConfig.APIURL,
		config.SAMSClientConfig.ConnConfig.ExternalURL, // default
	)
	samsClient := sams.NewClient(
		samsAPIURL,
		clientcredentials.Config{
			ClientID:     config.SAMSClientConfig.ClientID,
			ClientSecret: config.SAMSClientConfig.ClientSecret,
			// Since we are only using our SAMS client to verify supplied token,
			// we just issue tokens with a minimal set of scopes.
			Scopes:   []string{"openid", "profile", "email"},
			TokenURL: fmt.Sprintf("%s/oauth/token", samsAPIURL),
		})

	return newMaintenanceHandler(logger, next, redisKV, samsClient)
}

// Implements NewMaintenanceHandler, but allowing for the SAMS client to be provided to
// make testing easier.
func newMaintenanceHandler(
	logger log.Logger, next http.Handler, redisKV redispool.KeyValue,
	samsClient sams.Client) http.Handler {
	samsAuther := sams.Authenticator{
		Logger:     logger,
		SAMSClient: samsClient,
	}
	mHandlers := &maintenanceHandlers{
		logger: logger,
		redis:  redisKV,
	}

	// Create an HTTP router specific to these endpoints, to make registration easier.
	router := mux.NewRouter().PathPrefix("/maintenance/").Subrouter()
	registerMaintenanceHandler := func(method, url string, h http.HandlerFunc, requiredScope sams.Scope) {
		// Wrap the maintenance handler with a few of the standard Cody Gateway middleware functions.
		wrappedHandler := instrumentation.HTTPMiddleware(
			"diagnostics",
			requestlogger.Middleware(
				logger,
				samsAuther.RequireScopes([]sams.Scope{requiredScope}, h)),
			otelhttp.WithPublicEndpoint())

		// Finally, attach it to the router.
		router.
			Handle(url, wrappedHandler).
			Methods(method)
	}

	// Register the specific API endpoints.
	registerMaintenanceHandler(http.MethodGet, "/flagged-requests", mHandlers.ListFlaggedPrompts, samsScopeFlaggedPromptRead)
	registerMaintenanceHandler(http.MethodGet, "/flagged-requests/{flaggedPromptKey}", mHandlers.GetFlaggedPrompt, samsScopeFlaggedPromptRead)

	// Return the http.Handler. Yes, this is a horrible design wart, since we are using an
	// "HTTP middleware" pattern, but not actually registering middleware. To address this,
	// we'd need to rework Cody Gateway's httpapi.go and friends to pass around the mux.Router,
	// and register middleware via mux.Router::Use and attach endpoints directly.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/maintenance/") {
			router.ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

type maintenanceHandlers struct {
	logger log.Logger
	redis  redispool.KeyValue
}

type flaggedPromptInfo struct {
	Key     string `json:"key"`
	TraceID string `json:"traceId"`
	Feature string `json:"feature"`
	UserID  string `json:"userId"`
}

func parseFlaggedPromptKey(key string) (flaggedPromptInfo, error) {
	shortKey := strings.TrimPrefix(key, "prompt:")
	parts := strings.Split(shortKey, ":")
	if len(parts) != 3 {
		return flaggedPromptInfo{}, errors.Newf("invalid prompt key: %s", key)
	}

	return flaggedPromptInfo{
		Key:     key,
		TraceID: parts[0],
		Feature: parts[1],
		UserID:  parts[2],
	}, nil
}

type listFlaggedPromptsResponse struct {
	Prompts []flaggedPromptInfo `json:"prompts"`
}

// ListFlaggedPrompts returns summary information about all flagged requests currently sitting in our
// short-term memory. All of these keys will be deleted by the Redis TTL.
func (mh *maintenanceHandlers) ListFlaggedPrompts(w http.ResponseWriter, r *http.Request) {
	mh.logger.Info("listing flagged prompts")

	flaggedPrompts, err := mh.redis.Keys("prompt:*")
	if err != nil {
		mh.logger.Error("listing flagged prompts", log.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var resp listFlaggedPromptsResponse
	for _, flaggedPromptKey := range flaggedPrompts {
		promptInfo, err := parseFlaggedPromptKey(flaggedPromptKey)
		if err != nil {
			mh.logger.Error("error parsing flagged prompt Redis key", log.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		resp.Prompts = append(resp.Prompts, promptInfo)
	}

	// Write the response.
	respJSON, err := json.Marshal(&resp)
	if err != nil {
		mh.logger.Error("marshalling response object", log.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(respJSON); err != nil {
		mh.logger.Error("writing HTTP response", log.Error(err))
	}
}

// GetFlaggedPrompt returns the flagged prompt, identified by the route parameter.
func (mh *maintenanceHandlers) GetFlaggedPrompt(w http.ResponseWriter, r *http.Request) {
	routeVars := mux.Vars(r)
	flaggedPromptKey, ok := routeVars["flaggedPromptKey"]
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	mh.logger.Info("getting flagged prompts", log.String("key", flaggedPromptKey))

	if !strings.HasPrefix(flaggedPromptKey, "prompt:") {
		http.Error(w, "Invalid prompt key", http.StatusBadRequest)
		return
	}

	// Lookup the flagged prompt. It's possible the prompt was TTL'd and no longer exists.
	promptValue := mh.redis.Get(flaggedPromptKey)
	if promptValue.IsNil() {
		http.Error(w, "Prompt not found", http.StatusNotFound)
		return
	}

	prompt, err := promptValue.String()
	if err != nil {
		mh.logger.Error("reading flagged prompt", log.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Write the prompt.
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write([]byte(prompt)); err != nil {
		mh.logger.Error("writing HTTP response", log.Error(err))
	}
}
