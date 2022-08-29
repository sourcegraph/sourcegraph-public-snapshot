package codeintel

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/httpapi"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores"
	store "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifuploadstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	uploadshttp "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/transport/http"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type Services struct {
	dbStore *store.Store

	// shared with executor queue
	InternalUploadHandler http.Handler
	ExternalUploadHandler http.Handler

	gitserverClient *gitserver.Client

	// used by resolvers
	AutoIndexingSvc *autoindexing.Service
	UploadsSvc      *uploads.Service
	CodeNavSvc      *codenav.Service
	PoliciesSvc     *policies.Service
	UploadSvc       *uploads.Service
}

func NewServices(ctx context.Context, config *Config, siteConfig conftypes.WatchableSiteConfig, db database.DB) (*Services, error) {
	// Initialize tracing/metrics
	logger := log.Scoped("codeintel", "codeintel services")
	observationContext := &observation.Context{
		Logger:     logger,
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	// Initialize stores
	dbStore := store.NewWithDB(db, observationContext)

	// Connect to the separate LSIF database
	codeIntelDBConnection := mustInitializeCodeIntelDB(logger)

	// Initialize lsif stores (TODO: these should be integrated, they are basically pointing to the same thing)
	codeIntelLsifStore := database.NewDBWith(observationContext.Logger, codeIntelDBConnection)
	uploadStore, err := lsifuploadstore.New(context.Background(), config.LSIFUploadStoreConfig, observationContext)
	if err != nil {
		logger.Fatal("Failed to initialize upload store", log.Error(err))
	}

	// Initialize gitserver client & repoupdater
	gitserverClient := gitserver.New(db, dbStore, observationContext)
	repoUpdaterClient := repoupdater.New(observationContext)

	// Initialize services
	uploadSvc := uploads.GetService(db, codeIntelLsifStore, gitserverClient)
	codenavSvc := codenav.GetService(db, codeIntelLsifStore, uploadSvc, gitserverClient)
	policySvc := policies.GetService(db, uploadSvc, gitserverClient)
	autoindexingSvc := autoindexing.GetService(db, uploadSvc, gitserverClient, repoUpdaterClient)

	// Initialize http endpoints
	operations := httpapi.NewOperations(observationContext)
	newUploadHandler := func(internal bool) http.Handler {
		if false {
			// Until this handler has been implemented, we retain the origial
			// LSIF update handler.
			//
			// See https://github.com/sourcegraph/sourcegraph/issues/33375

			return uploadshttp.GetHandler(uploadSvc)
		}

		return httpapi.NewUploadHandler(
			db,
			&httpapi.DBStoreShim{Store: dbStore},
			uploadStore,
			internal,
			httpapi.DefaultValidatorByCodeHost,
			operations,
		)
	}
	internalUploadHandler := newUploadHandler(true)
	externalUploadHandler := newUploadHandler(false)

	return &Services{
		dbStore: dbStore,

		InternalUploadHandler: internalUploadHandler,
		ExternalUploadHandler: externalUploadHandler,

		gitserverClient: gitserverClient,

		AutoIndexingSvc: autoindexingSvc,
		UploadsSvc:      uploadSvc,
		CodeNavSvc:      codenavSvc,
		PoliciesSvc:     policySvc,
		UploadSvc:       uploadSvc,
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
