package shared

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/redigo"
	"github.com/gomodule/redigo/redis"
	"github.com/slack-go/slack"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/notify"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/requestclient"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/actor/anonymous"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/actor/dotcomuser"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/actor/productsubscription"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/httpapi"
)

func Main(ctx context.Context, obctx *observation.Context, ready service.ReadyFunc, config *Config) error {
	// Enable tracing, at this point tracing wouldn't have been enabled yet because
	// we run Cody Gateway without conf which means Sourcegraph tracing is not enabled.
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
			return errors.Wrap(err, "create BigQuery event logger")
		}

		// If a buffer is configured, wrap in events.BufferedLogger
		if config.BigQuery.EventBufferSize > 0 {
			eventLogger = events.NewBufferedLogger(obctx.Logger, eventLogger, config.BigQuery.EventBufferSize)
		}
	} else {
		eventLogger = events.NewStdoutLogger(obctx.Logger)

		// Useful for testing event logging in a way that has latency that is
		// somewhat similar to BigQuery.
		if os.Getenv("CODY_GATEWAY_BUFFERED_LAGGY_EVENT_LOGGING_FUN_TIMES_MODE") == "true" {
			eventLogger = events.NewBufferedLogger(
				obctx.Logger,
				events.NewDelayedLogger(eventLogger),
				config.BigQuery.EventBufferSize)
		}
	}

	dotcomClient := dotcom.NewClient(config.Dotcom.URL, config.Dotcom.AccessToken)

	// Supported actor/auth sources
	sources := actor.NewSources(
		anonymous.NewSource(config.AllowAnonymous, config.ActorConcurrencyLimit),
		productsubscription.NewSource(
			obctx.Logger,
			rcache.NewWithTTL(fmt.Sprintf("product-subscriptions:%s", productsubscription.SourceVersion), int(config.SourcesCacheTTL.Seconds())),
			dotcomClient,
			config.Dotcom.InternalMode,
			config.ActorConcurrencyLimit,
		),
		dotcomuser.NewSource(obctx.Logger,
			rcache.NewWithTTL(fmt.Sprintf("dotcom-users:%s", dotcomuser.SourceVersion), int(config.SourcesCacheTTL.Seconds())),
			dotcomClient,
			config.ActorConcurrencyLimit,
		),
	)

	authr := &auth.Authenticator{
		Logger:      obctx.Logger.Scoped("auth", "authentication middleware"),
		EventLogger: eventLogger,
		Sources:     sources,
	}

	rs := newRedisStore(redispool.Cache)

	// Ignore the error because it's already validated in the config.
	dotcomURL, _ := url.Parse(config.Dotcom.URL)
	dotcomURL.Path = ""
	rateLimitNotifier := notify.NewSlackRateLimitNotifier(
		obctx.Logger,
		redispool.Cache,
		dotcomURL.String(),
		config.ActorRateLimitNotify.Thresholds,
		config.ActorRateLimitNotify.SlackWebhookURL,
		slack.PostWebhookContext,
	)

	// Set up our handler chain, which is run from the bottom up. Application handlers
	// come last.
	handler := httpapi.NewHandler(obctx.Logger, eventLogger, rs, authr, &httpapi.Config{
		RateLimitNotifier:          rateLimitNotifier,
		AnthropicAccessToken:       config.Anthropic.AccessToken,
		AnthropicAllowedModels:     config.Anthropic.AllowedModels,
		AnthropicMaxTokensToSample: config.Anthropic.MaxTokensToSample,
		OpenAIAccessToken:          config.OpenAI.AccessToken,
		OpenAIOrgID:                config.OpenAI.OrgID,
		OpenAIAllowedModels:        config.OpenAI.AllowedModels,
		EmbeddingsAllowedModels:    config.AllowedEmbeddingsModels,
	})

	// Diagnostic layers
	handler = httpapi.NewDiagnosticsHandler(obctx.Logger, handler, config.DiagnosticsSecret, sources)

	// Collect request client for downstream handlers. Outside of dev, we always set up
	// Cloudflare in from of Cody Gateway. This comes first.
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

	// Collect background routines
	backgroundRoutines := []goroutine.BackgroundRoutine{
		server,
		sources.Worker(obctx, sourceWorkerMutex, config.SourcesSyncInterval),
	}
	if w, ok := eventLogger.(goroutine.BackgroundRoutine); ok {
		// eventLogger is events.BufferedLogger
		backgroundRoutines = append(backgroundRoutines, w)
	}
	// Block until done
	goroutine.MonitorBackgroundRoutines(ctx, backgroundRoutines...)

	return nil
}

func newRedisStore(store redispool.KeyValue) limiter.RedisStore {
	return &redisStore{
		store: store,
	}
}

type redisStore struct {
	store redispool.KeyValue
}

func (s *redisStore) Incrby(key string, val int) (int, error) {
	return s.store.Incrby(key, val)
}

func (s *redisStore) GetInt(key string) (int, error) {
	i, err := s.store.Get(key).Int()
	if err != nil && err != redis.ErrNil {
		return 0, err
	}
	return i, nil
}

func (s *redisStore) TTL(key string) (int, error) {
	return s.store.TTL(key)
}

func (s *redisStore) Expire(key string, ttlSeconds int) error {
	return s.store.Expire(key, ttlSeconds)
}
