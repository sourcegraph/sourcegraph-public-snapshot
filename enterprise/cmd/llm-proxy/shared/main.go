package shared

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/gorilla/mux"
	"github.com/gregjones/httpcache"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/instrumentation"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/actor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/actor/productsubscription"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/limiter"
)

func Main(ctx context.Context, obctx *observation.Context, ready service.ReadyFunc, config *Config) error {
	// Enable tracing, at this point tracing wouldn't have been enabled yet because
	// we run LLM-proxy without conf which means Sourcegraph tracing is not enabled.
	shutdownTracing, err := maybeEnableTracing(ctx, obctx.Logger.Scoped("tracing", "tracing configuration"))
	if err != nil {
		return errors.Wrap(err, "maybeEnableTracing")
	}
	defer shutdownTracing()

	handler := newServiceHandler(obctx.Logger, config)
	handler = rateLimit(obctx.Logger, redispool.Cache, handler)
	handler = authenticate(
		obctx.Logger,
		rcache.New("llm-proxy-tokens"),
		dotcom.NewClient(config.Dotcom.AccessToken),
		handler,
		authenticateOptions{
			AllowAnonymous: config.AllowAnonymous,
		},
	)
	handler = instrumentation.HTTPMiddleware("llm-proxy", handler)

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

func newServiceHandler(logger log.Logger, config *Config) http.Handler {
	r := mux.NewRouter()

	// For cluster liveness and readiness probes
	healthzLogger := logger.Scoped("healthz", "healthz checks")
	r.HandleFunc("/-/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := healthz(r.Context()); err != nil {
			healthzLogger.Error("check failed", log.Error(err))

			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("healthz: " + err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("healthz: ok"))
	})
	r.HandleFunc("/-/__version", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(version.Version()))
		return
	})

	// V1 service routes
	v1router := r.PathPrefix("/v1").Subrouter()
	v1router.Handle("/completions/anthropic",
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

			w.WriteHeader(resp.StatusCode)
			_ = resp.Header.Write(w)

			_, _ = io.Copy(w, resp.Body)
		}),
	).Methods(http.MethodPost)

	return r
}

type authenticateOptions struct {
	// TODO: Later maybe make this a configurable rate limit as well.
	AllowAnonymous bool
}

func authenticate(logger log.Logger, cache httpcache.Cache, dotComClient graphql.Client, next http.Handler, opts authenticateOptions) http.Handler {
	productSubscriptions := productsubscription.NewSource(logger, cache, dotComClient)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token == "" {
			// Anonymous requests. Based on configuration, either allow or disallow.
			if opts.AllowAnonymous {
				// TODO: We need to evaluate these to an actor type as well to rate-limit
				// anonymous users.
				next.ServeHTTP(w, r)
				return
			}

			responseJSONError(logger, w, http.StatusUnauthorized,
				errors.New("unauthorized access is not allowed"))
			return
		}

		act, err := productSubscriptions.Get(r.Context(), token)
		if err != nil {
			responseJSONError(logger, w, http.StatusUnauthorized, err)
			return
		}

		if !act.AccessEnabled {
			responseJSONError(logger, w, http.StatusForbidden,
				errors.New("license archived or LLM proxy access not enabled"))
			return
		}

		r = r.WithContext(actor.WithActor(r.Context(), act))
		next.ServeHTTP(w, r)
	})
}

func rateLimit(logger log.Logger, cache limiter.RedisStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := actor.FromContext(r.Context()).Limiter(cache)

		err := l.TryAcquire(r.Context())

		var rateLimitExceeded limiter.RateLimitExceededError
		if err != nil {
			if errors.As(err, &rateLimitExceeded) {
				rateLimitExceeded.WriteResponse(w)
				return
			}

			responseJSONError(logger, w, http.StatusInternalServerError, err)
		}

		next.ServeHTTP(w, r)
	})
}

func healthz(ctx context.Context) error {
	// Check redis health
	rpool, ok := redispool.Cache.Pool()
	if !ok {
		return errors.New("redis: not available")
	}
	rconn, err := rpool.GetContext(ctx)
	if err != nil {
		return errors.Wrap(err, "redis: failed to get conn")
	}
	defer rconn.Close()

	data, err := rconn.Do("PING")
	if err != nil {
		return errors.Wrap(err, "redis: failed to ping")
	}
	if data != "PONG" {
		return errors.New("redis: failed to ping: no pong received")
	}

	return nil
}
