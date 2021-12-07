package codeintel

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/httpapi"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindex/enqueuer"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/repoupdater"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/connections"
	"github.com/sourcegraph/sourcegraph/internal/database/locker"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
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
	indexEnqueuer   *enqueuer.IndexEnqueuer
}

func NewServices(ctx context.Context, siteConfig conftypes.SiteConfigQuerier, db database.DB) (*Services, error) {
	if err := config.UploadStoreConfig.Validate(); err != nil {
		return nil, errors.Errorf("failed to load config: %s", err)
	}

	// Initialize tracing/metrics
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	// Connect to database
	codeIntelDB := mustInitializeCodeIntelDB()

	// Initialize stores
	dbStore := store.NewWithDB(db, observationContext)
	locker := locker.NewWithDB(db, "codeintel")
	lsifStore := lsifstore.NewStore(codeIntelDB, siteConfig, observationContext)
	uploadStore, err := uploadstore.CreateLazy(context.Background(), config.UploadStoreConfig, observationContext)
	if err != nil {
		log.Fatalf("Failed to initialize upload store: %s", err)
	}

	// Initialize http endpoints
	operations := httpapi.NewOperations(observationContext)
	newUploadHandler := func(internal bool) http.Handler {
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

	// Initialize gitserver client
	gitserverClient := gitserver.New(dbStore, observationContext)
	repoUpdaterClient := repoupdater.New(observationContext)

	// Initialize the index enqueuer
	indexEnqueuer := enqueuer.NewIndexEnqueuer(&enqueuer.DBStoreShim{Store: dbStore}, gitserverClient, repoUpdaterClient, config.AutoIndexEnqueuerConfig, observationContext)

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
	}, nil
}

func mustInitializeCodeIntelDB() *sql.DB {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.CodeIntelPostgresDSN
	})
	db, err := connections.NewCodeIntelDB(dsn, "frontend", true)
	if err != nil {
		log.Fatalf("Failed to connect to codeintel database: %s", err)
	}
	return db
}
