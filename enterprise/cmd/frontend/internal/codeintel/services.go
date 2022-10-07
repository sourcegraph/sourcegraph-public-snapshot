package codeintel

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifuploadstore"
	uploadshttp "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/transport/http"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type FrontendServices struct {
	codeintel.Services
	gitserverClient  *gitserver.Client
	NewUploadHandler func(withCodeHostAuth bool) http.Handler
}

func NewServices(ctx context.Context, config *Config, siteConfig conftypes.WatchableSiteConfig, db database.DB) (*FrontendServices, error) {
	logger := log.Scoped("codeintel", "codeintel services")

	// Connect to the separate LSIF database
	codeIntelDB := mustInitializeCodeIntelDB(logger)

	services, err := codeintel.GetServices(codeintel.Databases{
		DB:          db,
		CodeIntelDB: codeIntelDB,
	})
	if err != nil {
		return nil, err
	}

	// Initialize blob stores
	observationContext := &observation.Context{
		Logger:     logger,
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}
	uploadStore, err := lsifuploadstore.New(context.Background(), config.LSIFUploadStoreConfig, observationContext)
	if err != nil {
		return nil, err
	}

	return &FrontendServices{
		Services:        services,
		gitserverClient: gitserver.New(db, &observation.TestContext),
		NewUploadHandler: func(withCodeHostAuth bool) http.Handler {
			return uploadshttp.GetHandler(services.UploadsService, db, uploadStore, withCodeHostAuth)
		},
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
