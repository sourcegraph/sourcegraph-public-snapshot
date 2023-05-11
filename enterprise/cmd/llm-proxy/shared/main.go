package shared

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/redigo"

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
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/actor/anonymous"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/actor/productsubscription"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/limiter"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/response"
)

func Main(ctx context.Context, obctx *observation.Context, ready service.ReadyFunc, config *Config) error {
	// Enable tracing, at this point tracing wouldn't have been enabled yet because
	// we run LLM-proxy without conf which means Sourcegraph tracing is not enabled.
	shutdownTracing, err := maybeEnableTracing(ctx, obctx.Logger.Scoped("tracing", "tracing configuration"))
	if err != nil {
		return errors.Wrap(err, "maybeEnableTracing")
	}
	defer shutdownTracing()

	// Supported actor/auth sources
	sources := actor.Sources{
		anonymous.NewSource(config.AllowAnonymous),
		productsubscription.NewSource(
			obctx.Logger,
			rcache.New("product-subscriptions"),
			dotcom.NewClient(config.Dotcom.URL, config.Dotcom.AccessToken)),
	}

	// Set up our handler chain, which is run from the bottom up
	handler := newServiceHandler(obctx.Logger, config)
	handler = rateLimit(obctx.Logger, newPrefixRedisStore("rate_limit:", redispool.Cache), handler)
	handler = &auth.Authenticator{
		Log:     obctx.Logger.Scoped("auth", "authentication middleware"),
		Sources: sources,
		Next:    handler,
	}
	handler = instrumentation.HTTPMiddleware("llm-proxy", handler)
	handler = httpLogger(obctx.Logger.Scoped("httpAPI", ""), handler)

	// Initialize our server
	server := httpserver.NewFromAddr(config.Address, &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Handler:      handler,
	})

	// Set up redis-based distributed mutex for the source syncer worker
	p, ok := redispool.Store.Pool()
	if !ok {
		return errors.New("real redis is required")
	}
	sourceWorkerMutex := redsync.New(redigo.NewPool(p)).NewMutex("source-syncer-worker",
		// Do not retry endlessly becuase it's very likely that someone else has
		// a long-standing hold on the mutex. We will try again on the next periodic
		// goroutine run.
		redsync.WithTries(1),
		// Expire locks at 2x sync interval to avoid contention while avoiding
		// the lock getting stuck for too long if something happens. Every handler
		// iteration, we will extend the lock.
		redsync.WithExpiry(2*config.SourcesSyncInterval))

	// Mark health server as ready and go!
	ready()
	obctx.Logger.Info("service ready", log.String("address", config.Address))

	// Block until done
	goroutine.MonitorBackgroundRoutines(ctx,
		server,
		sources.Worker(obctx.Logger, sourceWorkerMutex, config.SourcesSyncInterval),
	)

	return nil
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
	})

	// V1 service routes
	v1router := r.PathPrefix("/v1").Subrouter()
	v1router.Handle("/completions/anthropic",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ar, err := http.NewRequest(http.MethodPost, "https://api.anthropic.com/v1/complete", r.Body)
			if err != nil {
				response.JSONError(logger, w, http.StatusInternalServerError, errors.Errorf("failed to create request: %s", err))
				return
			}

			// Mimic headers set by the official Anthropic client:
			// https://sourcegraph.com/github.com/anthropics/anthropic-sdk-typescript@493075d70f50f1568a276ed0cb177e297f5fef9f/-/blob/src/index.ts
			ar.Header.Set("Cache-Control", "no-cache")
			ar.Header.Set("Accept", "application/json")
			ar.Header.Set("Content-Type", "application/json")
			ar.Header.Set("Client", "sourcegraph-llm-proxy/1.0")
			ar.Header.Set("X-API-Key", config.Anthropic.AccessToken)

			resp, err := httpcli.ExternalDoer.Do(ar)
			if err != nil {
				response.JSONError(logger, w, http.StatusInternalServerError, errors.Errorf("failed to make request to Anthropic: %s", err))
				return
			}
			defer func() { _ = resp.Body.Close() }()

			for k, vv := range resp.Header {
				for _, v := range vv {
					w.Header().Add(k, v)
				}
			}
			w.WriteHeader(resp.StatusCode)

			_, _ = io.Copy(w, resp.Body)
		}),
	).Methods(http.MethodPost)

	return r
}

func rateLimit(logger log.Logger, cache limiter.RedisStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := actor.FromContext(r.Context()).Limiter(cache)

		err := l.TryAcquire(r.Context())

		if err != nil {
			var rateLimitExceeded limiter.RateLimitExceededError
			if errors.As(err, &rateLimitExceeded) {
				rateLimitExceeded.WriteResponse(w)
				return
			}

			if errors.Is(err, limiter.NoAccessError{}) {
				response.JSONError(logger, w, http.StatusForbidden, err)
				return
			}

			response.JSONError(logger, w, http.StatusInternalServerError, err)
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

func httpLogger(logger log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("Request", log.String("method", r.Method), log.String("path", r.URL.Path))
		next.ServeHTTP(w, r)
	})
}

func newPrefixRedisStore(prefix string, store redispool.KeyValue) limiter.RedisStore {
	return &prefixRedisStore{
		prefix: prefix,
		store:  store,
	}
}

type prefixRedisStore struct {
	prefix string
	store  redispool.KeyValue
}

func (s *prefixRedisStore) Incr(key string) (int, error) {
	return s.store.Incr(s.prefix + key)
}

func (s *prefixRedisStore) TTL(key string) (int, error) {
	return s.store.TTL(s.prefix + key)
}

func (s *prefixRedisStore) Expire(key string, ttlSeconds int) error {
	return s.store.Expire(s.prefix+key, ttlSeconds)
}
