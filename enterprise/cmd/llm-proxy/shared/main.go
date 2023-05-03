package shared

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/actor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/instrumentation"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func Main(ctx context.Context, obctx *observation.Context, ready service.ReadyFunc, config *Config) error {
	handler := newHandler(obctx.Logger, config)
	handler = authenticate(obctx.Logger, config.Dotcom.AccessToken, handler)
	handler = trace.HTTPMiddleware(obctx.Logger, handler, conf.DefaultClient())
	handler = instrumentation.HTTPMiddleware("", handler)

	server := httpserver.NewFromAddr(config.Address, &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Handler:      handler,
	})

	// Mark health server as ready and go!
	ready()
	obctx.Logger.Info("service ready", log.String("address", config.Address))

	goroutine.MonitorBackgroundRoutines(ctx, server)

	return nil
}

func responseJSONError(logger log.Logger, w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	err = json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
	})
	if err != nil {
		logger.Error("failed to write response", log.Error(err))
	}
}

func newHandler(logger log.Logger, config *Config) http.Handler {
	r := mux.NewRouter()
	s := r.PathPrefix("/v1").Subrouter()

	s.Handle("/completions/anthropic",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r, err := http.NewRequest(http.MethodPost, "https://api.anthropic.com/v1/complete", r.Body)
			if err != nil {
				responseJSONError(logger, w, http.StatusInternalServerError, errors.Errorf("failed to create request: %s", err))
				return
			}

			// Mimic headers set by the official Anthropic client:
			// https://sourcegraph.com/github.com/anthropics/anthropic-sdk-typescript@493075d70f50f1568a276ed0cb177e297f5fef9f/-/blob/src/index.ts
			r.Header.Set("Cache-Control", "no-cache")
			r.Header.Set("Accept", "application/json")
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Client", "sourcegraph-llm-proxy/1.0")
			r.Header.Set("X-API-Key", config.Anthropic.AccessToken)

			resp, err := httpcli.ExternalDoer.Do(r)
			if err != nil {
				responseJSONError(logger, w, http.StatusInternalServerError, errors.Errorf("failed to make request to Anthropic: %s", err))
				return
			}
			defer func() { _ = resp.Body.Close() }()

			_, _ = io.Copy(w, resp.Body)
		}),
	).Methods(http.MethodPost)
	return r
}

func authenticate(logger log.Logger, dotcomAccessToken string, next http.Handler) http.Handler {
	dotcomClient := dotcom.NewClient(dotcomAccessToken)
	tokenCache := rcache.New("llm-proxy-token")
	getSubscriptionFromCache := func(token string) (_ *actor.Subscription, hit bool) {
		data, hit := tokenCache.Get(token)
		if !hit {
			return nil, false
		}

		var subscription actor.Subscription
		if err := json.Unmarshal(data, &subscription); err != nil {
			logger.Error("failed to unmarshal subscription", log.Error(err))
			return nil, false
		}
		return &subscription, true
	}
	saveSubscriptionToCache := func(token string, subscription *actor.Subscription) {
		data, err := json.Marshal(subscription)
		if err != nil {
			logger.Error("failed to marshal subscription", log.Error(err))
			return
		}

		tokenCache.Set(token, data)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token == "" {
			// Anonymous requests
			next.ServeHTTP(w, r)
			return
		}

		subscription, hit := getSubscriptionFromCache(token)
		if !hit {
			resp, err := dotcom.CheckAccessToken(r.Context(), dotcomClient, token)
			if err != nil {
				responseJSONError(logger, w, http.StatusUnauthorized, errors.Errorf("failed to check access token: %s", err))
				return
			}

			if resp.Dotcom.ProductSubscriptionByAccessToken.IsArchived ||
				!resp.Dotcom.ProductSubscriptionByAccessToken.LlmProxyAccess.Enabled {
				responseJSONError(logger, w, http.StatusBadRequest, errors.New("license archived or LLM proxy access not enabled"))
				return
			}

			subscription = &actor.Subscription{
				ID:        resp.Dotcom.ProductSubscriptionByAccessToken.Id,
				RateLimit: resp.Dotcom.ProductSubscriptionByAccessToken.LlmProxyAccess.RateLimit,
			}
			saveSubscriptionToCache(token, subscription)
		}

		r = r.WithContext(actor.WithActor(r.Context(), &actor.Actor{Subscription: subscription}))
		next.ServeHTTP(w, r)
	})
}
