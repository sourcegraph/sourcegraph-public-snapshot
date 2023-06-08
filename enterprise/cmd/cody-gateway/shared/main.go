package shared

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/redigo"
	"github.com/gomodule/redigo/redis"
	"github.com/slack-go/slack"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/instrumentation"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/requestclient"
	"github.com/sourcegraph/sourcegraph/internal/service"
	sgtrace "github.com/sourcegraph/sourcegraph/internal/trace"
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
	sources := actor.Sources{
		anonymous.NewSource(config.AllowAnonymous, config.ActorConcurrencyLimit),
		productsubscription.NewSource(
			obctx.Logger,
			rcache.New("product-subscriptions"),
			dotcomClient,
			config.Dotcom.InternalMode,
			config.ActorConcurrencyLimit,
		),
		dotcomuser.NewSource(obctx.Logger,
			rcache.New("dotcom-users"),
			dotcomClient,
			config.ActorConcurrencyLimit,
		),
	}

	authr := &auth.Authenticator{
		Logger:      obctx.Logger.Scoped("auth", "authentication middleware"),
		EventLogger: eventLogger,
		Sources:     sources,
	}

	rs := newRedisStore(redispool.Cache)

	obctx.Logger.Debug("concurrency limit",
		log.Float32("percentage", config.ActorConcurrencyLimit.Percentage),
		log.String("internal", config.ActorConcurrencyLimit.Interval.String()),
	)
	// Set up our handler chain, which is run from the bottom up. Application handlers
	// come last.
	handler := httpapi.NewHandler(obctx.Logger, eventLogger, rs, authr, &httpapi.Config{
		RateLimitAlerter:        newRateLimitAlerter(obctx.Logger, redispool.Cache, config.Dotcom.URL, config.ActorRateLimitAlert, slack.PostWebhookContext),
		AnthropicAccessToken:    config.Anthropic.AccessToken,
		AnthropicAllowedModels:  config.Anthropic.AllowedModels,
		OpenAIAccessToken:       config.OpenAI.AccessToken,
		OpenAIOrgID:             config.OpenAI.OrgID,
		OpenAIAllowedModels:     config.OpenAI.AllowedModels,
		EmbeddingsAllowedModels: config.AllowedEmbeddingsModels,
	})

	// Diagnostic layers
	handler = httpapi.NewDiagnosticsHandler(obctx.Logger, handler, config.DiagnosticsSecret)

	// Instrumentation layers
	handler = requestLogger(obctx.Logger.Scoped("requests", "HTTP requests"), handler)
	var otelhttpOpts []otelhttp.Option
	if !config.InsecureDev {
		// Outside of dev, we're probably running as a standalone service, so treat
		// incoming spans as links
		otelhttpOpts = append(otelhttpOpts, otelhttp.WithPublicEndpoint())
	}
	handler = instrumentation.HTTPMiddleware("cody-gateway", handler, otelhttpOpts...)

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

func requestLogger(logger log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only requestclient is available at the point, actor middleware is later
		rc := requestclient.FromContext(r.Context())

		sgtrace.Logger(r.Context(), logger).Debug("Request",
			log.String("method", r.Method),
			log.String("path", r.URL.Path),
			log.String("requestclient.ip", rc.IP),
			log.String("requestclient.forwardedFor", rc.ForwardedFor))

		next.ServeHTTP(w, r)
	})
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

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// tryAcquireRedisLockOnce attempts to acquire a Redis-based lock with the given
// key in a single pass. The locking algorithm is based on
// https://redis.io/commands/setnx/ for resolving deadlocks.
//
// CAUTION: To avoid releasing someone else's lock, the duration of the entire
// operation should be well-below the lock timeout.
//
// TODO(jchen): Modularize this into the redispool package.
func tryAcquireRedisLockOnce(rs redispool.KeyValue, lockKey string, lockTimeout time.Duration) (acquired bool, release func(), _ error) {
	timeout := time.Now().Add(lockTimeout).UnixNano()
	// Add a random number to decrease the chance of multiple processes falsely
	// believing they have the lock at the same time.
	lockToken := fmt.Sprintf("%d,%d", timeout, seededRand.Intn(1e6))

	release = func() {
		// Best effort to check we're releasing the lock we think we have. Note that it
		// is still technically possible the lock token has changed between the GET and
		// DEL since these are two separate operations, i.e. when the current lock happen
		// to be expired at this very moment.
		get, _ := rs.Get(lockKey).String()
		if get == lockToken {
			_ = rs.Del(lockKey)
		}
	}

	set, err := rs.SetNx(lockKey, lockToken)
	if err != nil {
		return false, nil, err
	} else if set {
		return true, release, nil
	}

	// We didn't get the lock, but we can check if the lock is expired.
	currentLockToken, err := rs.Get(lockKey).String()
	if err == redis.ErrNil {
		// Someone else got the lock and released it already.
		return false, nil, nil
	} else if err != nil {
		return false, nil, err
	}

	currentTimeout, _ := strconv.ParseInt(strings.SplitN(currentLockToken, ",", 2)[0], 10, 64)
	if currentTimeout > time.Now().UnixNano() {
		// The lock is still valid.
		return false, nil, nil
	}

	// The lock has expired, try to acquire it.
	get, err := rs.GetSet(lockKey, lockToken).String()
	if err != nil {
		return false, nil, err
	} else if get != currentLockToken {
		// Someone else got the lock
		return false, nil, nil
	}

	// We got the lock.
	return true, release, nil
}

// newRateLimitAlerter returns a function that sends an alert to the given Slack
// webhook URL when the threshold is met.
func newRateLimitAlerter(
	logger log.Logger,
	rs redispool.KeyValue,
	dotcomAPIURL string,
	config codygateway.ActorRateLimitAlertConfig,
	slackSender func(ctx context.Context, url string, msg *slack.WebhookMessage) error,
) func(actor *actor.Actor, feature codygateway.Feature, usagePercentage float32) {
	// Ignore the error because it's already validated in the config.
	dotcomURL, _ := url.Parse(dotcomAPIURL)
	dotcomURL.Path = ""

	logger = logger.Scoped("usageRateLimitAlert", "alerts for usage rate limit approaching threshold")
	return func(actor *actor.Actor, feature codygateway.Feature, usagePercentage float32) {
		if usagePercentage < config.Threshold {
			return
		}

		lockKey := fmt.Sprintf("rate_limit:%s:alert:lock:%s", feature, actor.ID)
		acquired, release, err := tryAcquireRedisLockOnce(rs, lockKey, 30*time.Second)
		if err != nil {
			logger.Error("failed to acquire lock", log.Error(err))
			return
		} else if !acquired {
			return
		}
		defer release()

		key := fmt.Sprintf("rate_limit:%s:alert:%s", feature, actor.ID)
		get, err := rs.Get(key).String()
		if err != nil && err != redis.ErrNil {
			logger.Error("failed to get last alerted time", log.Error(err))
			return
		}

		lastAlerted, err := time.Parse(time.RFC3339, get)
		if err == nil && !time.Now().After(lastAlerted.Add(config.Interval)) {
			return // Still in the cooldown period
		}
		defer func() {
			err := rs.Set(key, time.Now().Format(time.RFC3339))
			if err != nil {
				logger.Error("failed to set last alerted time", log.Error(err))
			}
		}()

		if config.SlackWebhookURL == "" {
			logger.Debug("new usage alert",
				log.Object("actor",
					log.String("id", actor.ID),
					log.String("source", actor.Source.Name()),
				),
				log.String("feature", string(feature)),
				log.Int("usagePercentage", int(usagePercentage*100)),
			)
			return
		}

		var text string
		switch codygateway.ActorSource(actor.Source.Name()) {
		case codygateway.ActorSourceProductSubscription:
			text = fmt.Sprintf("The actor <%[1]s/site-admin/dotcom/product/subscriptions/%[2]s|%[2]s> from %q has exceeded *%d%%* of its rate limit quota for `%s`.",
				dotcomURL.String(), actor.ID, actor.Source.Name(), int(usagePercentage*100), feature)
		case codygateway.ActorSourceDotcomUser:
			text = fmt.Sprintf("The actor <%[1]s/users/%[2]s|%[2]s> from %q has exceeded *%d%%* of its rate limit quota for `%s`.",
				dotcomURL.String(), actor.ID, actor.Source.Name(), int(usagePercentage*100), feature)
		default:
			text = fmt.Sprintf("The actor `%s` from %q has exceeded *%d%%* of its rate limit quota for `%s`.",
				actor.ID, actor.Source.Name(), int(usagePercentage*100), feature)
		}

		// NOTE: The context timeout must below the lock timeout we set above (30 seconds
		// ) to make sure the lock doesn't expire when we release it, i.e. avoid
		// releasing someone else's lock.
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		err = slackSender(
			ctx,
			config.SlackWebhookURL,
			&slack.WebhookMessage{
				Blocks: &slack.Blocks{
					BlockSet: []slack.Block{
						slack.NewSectionBlock(
							slack.NewTextBlockObject("mrkdwn", text, false, false),
							nil,
							nil,
						),
					},
				},
			},
		)
		if err != nil {
			logger.Error("failed to send Slack webhook", log.Error(err))
		}
	}
}
