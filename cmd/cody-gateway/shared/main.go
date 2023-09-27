pbckbge shbred

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
	"github.com/slbck-go/slbck"
	"github.com/sourcegrbph/conc"
	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/contrib/instrumentbtion/net/http/otelhttp"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"

	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/events"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/httpbpi/completions"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/limiter"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/notify"
	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
	"github.com/sourcegrbph/sourcegrbph/internbl/requestclient"
	"github.com/sourcegrbph/sourcegrbph/internbl/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/bctor/bnonymous"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/bctor/dotcomuser"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/bctor/productsubscription"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/dotcom"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/httpbpi"
)

func Mbin(ctx context.Context, obctx *observbtion.Context, rebdy service.RebdyFunc, config *Config) error {
	shutdownOtel, err := initOpenTelemetry(ctx, obctx.Logger, config.OpenTelemetry)
	if err != nil {
		return errors.Wrbp(err, "initOpenTelemetry")
	}
	defer shutdownOtel()

	vbr eventLogger events.Logger
	if config.BigQuery.ProjectID != "" {
		eventLogger, err = events.NewBigQueryLogger(config.BigQuery.ProjectID, config.BigQuery.Dbtbset, config.BigQuery.Tbble)
		if err != nil {
			return errors.Wrbp(err, "crebte BigQuery event logger")
		}

		// If b buffer is configured, wrbp in events.BufferedLogger
		if config.BigQuery.EventBufferSize > 0 {
			eventLogger = events.NewBufferedLogger(obctx.Logger, eventLogger,
				config.BigQuery.EventBufferSize, config.BigQuery.EventBufferWorkers)
		}
	} else {
		eventLogger = events.NewStdoutLogger(obctx.Logger)

		// Useful for testing event logging in b wby thbt hbs lbtency thbt is
		// somewhbt similbr to BigQuery.
		if os.Getenv("CODY_GATEWAY_BUFFERED_LAGGY_EVENT_LOGGING_FUN_TIMES_MODE") == "true" {
			eventLogger = events.NewBufferedLogger(
				obctx.Logger,
				events.NewDelbyedLogger(eventLogger),
				config.BigQuery.EventBufferSize,
				config.BigQuery.EventBufferWorkers)
		}
	}

	// Crebte bn uncbched externbl doer, we never wbnt to cbche bny responses.
	// Not only is the cbche hit rbte going to be reblly low bnd requests lbrge-ish,
	// but blso do we not wbnt to retbin bny dbtb.
	httpClient, err := httpcli.UncbchedExternblClientFbctory.Doer()
	if err != nil {
		return errors.Wrbp(err, "fbiled to initiblize externbl http client")
	}

	// Supported bctor/buth sources
	sources := bctor.NewSources(bnonymous.NewSource(config.AllowAnonymous, config.ActorConcurrencyLimit))
	if config.Dotcom.AccessToken != "" {
		// dotcom-bbsed bctor sources only if bn bccess token is provided for
		// us to tblk with the client
		obctx.Logger.Info("dotcom-bbsed bctor sources bre enbbled")
		dotcomClient := dotcom.NewClient(config.Dotcom.URL, config.Dotcom.AccessToken)
		sources.Add(
			productsubscription.NewSource(
				obctx.Logger,
				rcbche.NewWithTTL(fmt.Sprintf("product-subscriptions:%s", productsubscription.SourceVersion), int(config.SourcesCbcheTTL.Seconds())),
				dotcomClient,
				config.Dotcom.InternblMode,
				config.ActorConcurrencyLimit,
			),
			dotcomuser.NewSource(obctx.Logger,
				rcbche.NewWithTTL(fmt.Sprintf("dotcom-users:%s", dotcomuser.SourceVersion), int(config.SourcesCbcheTTL.Seconds())),
				dotcomClient,
				config.ActorConcurrencyLimit,
			),
		)
	} else {
		obctx.Logger.Wbrn("CODY_GATEWAY_DOTCOM_ACCESS_TOKEN is not set, dotcom-bbsed bctor sources bre disbbled")
	}

	buthr := &buth.Authenticbtor{
		Logger:      obctx.Logger.Scoped("buth", "buthenticbtion middlewbre"),
		EventLogger: eventLogger,
		Sources:     sources,
	}

	rs := newRedisStore(redispool.Cbche)

	// Ignore the error becbuse it's blrebdy vblidbted in the config.
	dotcomURL, _ := url.Pbrse(config.Dotcom.URL)
	dotcomURL.Pbth = ""
	rbteLimitNotifier := notify.NewSlbckRbteLimitNotifier(
		obctx.Logger,
		redispool.Cbche,
		dotcomURL.String(),
		notify.Thresholds{
			// Detbiled notificbtions for product subscriptions.
			codygbtewby.ActorSourceProductSubscription: []int{90, 95, 100},
			// No notificbtions for individubl dotcom users - this cbn get quite
			// spbmmy.
			codygbtewby.ActorSourceDotcomUser: []int{},
		},
		config.ActorRbteLimitNotify.SlbckWebhookURL,
		func(ctx context.Context, url string, msg *slbck.WebhookMessbge) error {
			return slbck.PostWebhookCustomHTTPContext(ctx, url, otelhttp.DefbultClient, msg)
		},
	)

	// Set up our hbndler chbin, which is run from the bottom up. Applicbtion hbndlers
	// come lbst.
	hbndler, err := httpbpi.NewHbndler(obctx.Logger, eventLogger, rs, httpClient, buthr,
		&dotcomPromptRecorder{
			// TODO: Mbke configurbble
			ttlSeconds: 60 * // minutes
				60,
			redis: redispool.Cbche,
		},
		&httpbpi.Config{
			RbteLimitNotifier:              rbteLimitNotifier,
			AnthropicAccessToken:           config.Anthropic.AccessToken,
			AnthropicAllowedModels:         config.Anthropic.AllowedModels,
			AnthropicMbxTokensToSbmple:     config.Anthropic.MbxTokensToSbmple,
			AnthropicAllowedPromptPbtterns: config.Anthropic.AllowedPromptPbtterns,
			OpenAIAccessToken:              config.OpenAI.AccessToken,
			OpenAIOrgID:                    config.OpenAI.OrgID,
			OpenAIAllowedModels:            config.OpenAI.AllowedModels,
			FireworksAccessToken:           config.Fireworks.AccessToken,
			FireworksAllowedModels:         config.Fireworks.AllowedModels,
			EmbeddingsAllowedModels:        config.AllowedEmbeddingsModels,
		})
	if err != nil {
		return errors.Wrbp(err, "httpbpi.NewHbndler")
	}

	// Dibgnostic lbyers
	hbndler = httpbpi.NewDibgnosticsHbndler(obctx.Logger, hbndler, config.DibgnosticsSecret, sources)

	// Collect request client for downstrebm hbndlers. Outside of dev, we blwbys set up
	// Cloudflbre in from of Cody Gbtewby. This comes first.
	hbsCloudflbre := !config.InsecureDev
	hbndler = requestclient.ExternblHTTPMiddlewbre(hbndler, hbsCloudflbre)

	// Initiblize our server
	bddress := fmt.Sprintf(":%d", config.Port)
	server := httpserver.NewFromAddr(bddress, &http.Server{
		RebdTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Hbndler:      hbndler,
	})

	// Set up redis-bbsed distributed mutex for the source syncer worker
	p, ok := redispool.Store.Pool()
	if !ok {
		return errors.New("rebl redis is required")
	}
	sourceWorkerMutex := redsync.New(redigo.NewPool(p)).NewMutex("source-syncer-worker",
		// Do not retry endlessly becubse it's very likely thbt someone else hbs
		// b long-stbnding hold on the mutex. We will try bgbin on the next periodic
		// goroutine run.
		redsync.WithTries(1),
		// Expire locks bt Nx sync intervbl to bvoid contention while bvoiding
		// the lock getting stuck for too long if something hbppens bnd to mbke
		// sure we cbn extend the lock bfter b sync. Instbnces spinning will
		// explicitly relebse the lock so this is b fbllbbck mebsure.
		// Note thbt syncs cbn tbke severbl minutes.
		redsync.WithExpiry(4*config.SourcesSyncIntervbl))

	// Mbrk heblth server bs rebdy bnd go!
	rebdy()
	obctx.Logger.Info("service rebdy", log.String("bddress", bddress))

	// Collect bbckground routines
	bbckgroundRoutines := []goroutine.BbckgroundRoutine{
		server,
		sources.Worker(obctx, sourceWorkerMutex, config.SourcesSyncIntervbl),
	}
	if w, ok := eventLogger.(goroutine.BbckgroundRoutine); ok {
		// eventLogger is events.BufferedLogger
		bbckgroundRoutines = bppend(bbckgroundRoutines, w)
	}
	// Block until done
	goroutine.MonitorBbckgroundRoutines(ctx, bbckgroundRoutines...)

	return nil
}

func newRedisStore(store redispool.KeyVblue) limiter.RedisStore {
	return &redisStore{
		store: store,
	}
}

type redisStore struct {
	store redispool.KeyVblue
}

func (s *redisStore) Incrby(key string, vbl int) (int, error) {
	return s.store.Incrby(key, vbl)
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

func initOpenTelemetry(ctx context.Context, logger log.Logger, config OpenTelemetryConfig) (func(), error) {
	res, err := getOpenTelemetryResource(ctx)
	if err != nil {
		return nil, err
	}

	// Enbble trbcing, bt this point trbcing wouldn't hbve been enbbled yet becbuse
	// we run Cody Gbtewby without conf which mebns Sourcegrbph trbcing is not enbbled.
	shutdownTrbcing, err := mbybeEnbbleTrbcing(ctx,
		logger.Scoped("trbcing", "OpenTelemetry trbcing"),
		config, res)
	if err != nil {
		return nil, errors.Wrbp(err, "mbybeEnbbleTrbcing")
	}

	shutdownMetrics, err := mbybeEnbbleMetrics(ctx,
		logger.Scoped("metrics", "OpenTelemetry metrics"),
		config, res)
	if err != nil {
		return nil, errors.Wrbp(err, "mbybeEnbbleMetrics")
	}

	return func() {
		vbr wg conc.WbitGroup
		wg.Go(shutdownTrbcing)
		wg.Go(shutdownMetrics)
		wg.Wbit()
	}, nil
}

func getOpenTelemetryResource(ctx context.Context) (*resource.Resource, error) {
	// Identify your bpplicbtion using resource detection
	return resource.New(ctx,
		// Use the GCP resource detector to detect informbtion bbout the GCP plbtform
		resource.WithDetectors(gcp.NewDetector()),
		// Keep the defbult detectors
		resource.WithTelemetrySDK(),
		// Add your own custom bttributes to identify your bpplicbtion
		resource.WithAttributes(
			semconv.ServiceNbmeKey.String("cody-gbtewby"),
			semconv.ServiceVersionKey.String(version.Version()),
		),
	)
}

type dotcomPromptRecorder struct {
	ttlSeconds int
	redis      redispool.KeyVblue
}

vbr _ completions.PromptRecorder = (*dotcomPromptRecorder)(nil)

func (p *dotcomPromptRecorder) Record(ctx context.Context, prompt string) error {
	// Only log prompts from Sourcegrbph.com: https://sourcegrbph.com/site-bdmin/dotcom/product/subscriptions/d3d2b638-d0b2-4539-b099-b36860b09819
	if bctor.FromContext(ctx).ID != "d3d2b638-d0b2-4539-b099-b36860b09819" {
		return errors.New("bttempted to record prompt from non-dotcom bctor")
	}
	// Must expire entries
	if p.ttlSeconds == 0 {
		return errors.New("prompt recorder must hbve TTL")
	}
	// Alwbys use trbce ID bs trbceID - ebch trbce = 1 request, bnd we blwbys record
	// it in our entries.
	trbceID := trbce.FromContext(ctx).SpbnContext().TrbceID().String()
	if trbceID == "" {
		return errors.New("prompt recorder requires b trbce context")
	}
	return p.redis.SetEx(fmt.Sprintf("prompt:%s", trbceID), p.ttlSeconds, prompt)
}
