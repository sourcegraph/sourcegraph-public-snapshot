package shared

import (
	"context"
	"net/http"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/redigo"
	"github.com/gomodule/redigo/redis"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/events"
	llmproxy "github.com/sourcegraph/sourcegraph/enterprise/internal/llm-proxy"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/instrumentation"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/requestclient"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/actor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/actor/anonymous"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/actor/productsubscription"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/httpapi"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/limiter"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/response"
)

func Main(ctx context.Context, obctx *observation.Context, ready service.ReadyFunc, config *Config) error {
	// Enable tracing, at this point tracing wouldn't have been enabled yet because
	// we run LLM-proxy without conf which means Sourcegraph tracing is not enabled.
	shutdownTracing, err := maybeEnableTracing(ctx,
		obctx.Logger.Scoped("tracing", "tracing configuration"),
		config.Trace)
	if err != nil {
		return errors.Wrap(err, "maybeEnableTracing")
	}
	defer shutdownTracing()

	var eventLogger events.Logger
	if config.BigQuery.ProjectID != "" {
		eventLogger, err = events.NewBigQueryLogger(config.BigQuery.ProjectID, config.BigQuery.Dataset, config.BigQuery.Table)
		if err != nil {
			return errors.Wrap(err, "create event logger")
		}
	} else {
		eventLogger = events.NewStdoutLogger(obctx.Logger)
	}

	// Supported actor/auth sources
	sources := actor.Sources{
		anonymous.NewSource(config.AllowAnonymous),
		productsubscription.NewSource(
			obctx.Logger,
			rcache.New("product-subscriptions"),
			dotcom.NewClient(config.Dotcom.URL, config.Dotcom.AccessToken),
			config.Dotcom.DevLicensesOnly),
	}

	// Set up our handler chain, which is run from the bottom up. Application handlers
	// come last.
	handler := httpapi.NewHandler(obctx.Logger, eventLogger, &httpapi.Config{
		AnthropicAccessToken: config.Anthropic.AccessToken,
		OpenAIAccessToken:    config.OpenAI.AccessToken,
		OpenAIOrgID:          config.OpenAI.OrgID,
	})

	// Authentication and rate-limting layers
	handler = rateLimit(obctx.Logger, eventLogger, newPrefixRedisStore("rate_limit:", redispool.Cache), handler)
	handler = &auth.Authenticator{
		Logger:      obctx.Logger.Scoped("auth", "authentication middleware"),
		EventLogger: eventLogger,
		Sources:     sources,
		Next:        handler,
	}

	// Instrumentation layers
	handler = httpLogger(obctx.Logger.Scoped("httpAPI", ""), handler)
	handler = instrumentation.HTTPMiddleware("llm-proxy", handler,
		otelhttp.WithPublicEndpoint())

	// Collect request client for downstream handlers. Outside of dev, we always set up
	// Cloudflare in from of LLM-proxy. This comes first.
	hasCloudflare := !config.InsecureDev
	handler = requestclient.ExternalHTTPMiddleware(handler, hasCloudflare)

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

func rateLimit(logger log.Logger, eventLogger events.Logger, cache limiter.RedisStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		act := actor.FromContext(r.Context())
		l := act.Limiter(cache)

		err := l.TryAcquire(r.Context())
		if err != nil {
			loggerErr := eventLogger.LogEvent(
				events.Event{
					Name:       llmproxy.EventNameRateLimited,
					Source:     act.Source.Name(),
					Identifier: act.ID,
					Metadata: map[string]any{
						"error": err.Error(),
					},
				},
			)
			if loggerErr != nil {
				logger.Error("failed to log event", log.Error(loggerErr))
			}

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
			return
		}

		next.ServeHTTP(w, r)
	})
}

func httpLogger(logger log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rc := requestclient.FromContext(r.Context())
		logger.Debug("Request",
			log.String("method", r.Method),
			log.String("path", r.URL.Path),
			log.String("requestclient.ip", rc.IP),
			log.String("requestclient.forwardedFor", rc.ForwardedFor))
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

func (s *prefixRedisStore) GetInt(key string) (int, error) {
	i, err := s.store.Get(s.prefix + key).Int()
	if err != nil && err != redis.ErrNil {
		return 0, err
	}
	return i, nil
}

func (s *prefixRedisStore) TTL(key string) (int, error) {
	return s.store.TTL(s.prefix + key)
}

func (s *prefixRedisStore) Expire(key string, ttlSeconds int) error {
	return s.store.Expire(s.prefix+key, ttlSeconds)
}
