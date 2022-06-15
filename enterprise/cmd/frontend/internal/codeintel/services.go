package codeintel

import (
	"context"
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/httpapi"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores"
	store "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifuploadstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	uploadshttp "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/transport/http"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/locker"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/sentry"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
)

type Services struct {
	dbStore     *store.Store
	lsifStore   *lsifstore.Store
	repoStore   database.RepoStore
	uploadStore uploadstore.Store

	// shared with executorqueue
	InternalUploadHandler http.Handler
	ExternalUploadHandler http.Handler

	locker          *locker.Locker
	gitserverClient *gitserver.Client
	indexEnqueuer   *autoindexing.Service
	hub             *sentry.Hub
}

func NewServices(ctx context.Context, config *Config, siteConfig conftypes.WatchableSiteConfig, db database.DB) (*Services, error) {
	// Initialize tracing/metrics
	logger := log.Scoped("codeintel", "codeintel services")
	observationContext := &observation.Context{
		Logger:     logger,
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	// Initialize sentry hub
	hub := mustInitializeSentryHub(logger, siteConfig)

	// Connect to database
	codeIntelDB := mustInitializeCodeIntelDB(logger)

	// Initialize stores
	dbStore := store.NewWithDB(db, observationContext)
	locker := locker.NewWith(db, "codeintel")
	lsifStore := lsifstore.NewStore(codeIntelDB, siteConfig, observationContext)
	uploadStore, err := lsifuploadstore.New(context.Background(), config.LSIFUploadStoreConfig, observationContext)
	if err != nil {
		logger.Fatal("Failed to initialize upload store", log.Error(err))
	}

	// Initialize http endpoints
	operations := httpapi.NewOperations(observationContext)
	newUploadHandler := func(internal bool) http.Handler {
		if false {
			// Until this handler has been implemented, we retain the origial
			// LSIF update handler.
			//
			// See https://github.com/sourcegraph/sourcegraph/issues/33375

			return uploadshttp.GetHandler(uploads.GetService(db))
		}

		return httpapi.NewUploadHandler(
			db,
			&httpapi.DBStoreShim{Store: dbStore},
			uploadStore,
			internal,
			httpapi.DefaultValidatorByCodeHost,
			operations,
			hub,
		)
	}
	internalUploadHandler := newUploadHandler(true)
	externalUploadHandler := newUploadHandler(false)

	// Initialize gitserver client
	gitserverClient := gitserver.New(db, dbStore, observationContext)
	repoUpdaterClient := repoupdater.New(observationContext)

	// Initialize the index enqueuer
	indexEnqueuer := autoindexing.GetService(db, &autoindexing.DBStoreShim{Store: dbStore}, gitserverClient, repoUpdaterClient)

	return &Services{
		dbStore:     dbStore,
		lsifStore:   lsifStore,
		repoStore:   database.ReposWith(dbStore.Store),
		uploadStore: uploadStore,

		InternalUploadHandler: internalUploadHandler,
		ExternalUploadHandler: externalUploadHandler,

		locker:          locker,
		gitserverClient: gitserverClient,
		indexEnqueuer:   indexEnqueuer,
		hub:             hub,
	}, nil
}

func mustInitializeCodeIntelDB(logger log.Logger) stores.CodeIntelDB {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.CodeIntelPostgresDSN
	})
	db, err := connections.EnsureNewCodeIntelDB(dsn, "frontend", &observation.TestContext)
	if err != nil {
		logger.Fatal("Failed to connect to codeintel database", log.Error(err))
	}
	return stores.NewCodeIntelDB(db)
}

func mustInitializeSentryHub(logger log.Logger, c conftypes.WatchableSiteConfig) *sentry.Hub {
	getDsn := func(c conftypes.SiteConfigQuerier) string {
		if c.SiteConfig().Log != nil && c.SiteConfig().Log.Sentry != nil {
			return c.SiteConfig().Log.Sentry.CodeIntelDSN
		}
		return ""
	}

	hub, err := sentry.NewWithDsn(getDsn(c), c, getDsn)
	if err != nil {
		logger.Fatal("Failed to initialize sentry hub", log.Error(err))
	}
	return hub
}
