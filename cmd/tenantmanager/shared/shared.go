// gitserver is the gitserver server.
package shared

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/tenantmanager/internal"
	"github.com/sourcegraph/sourcegraph/cmd/tenantmanager/internal/reconciler"
	"github.com/sourcegraph/sourcegraph/cmd/tenantmanager/internal/store"
	proto "github.com/sourcegraph/sourcegraph/cmd/tenantmanager/shared/v1"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/goroutine/recorder"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/grpc/grpcserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type LazyDebugserverEndpoint struct {
	tenantListEndpoint http.HandlerFunc
}

func Main(ctx context.Context, observationCtx *observation.Context, ready service.ReadyFunc, debugserverEndpoints *LazyDebugserverEndpoint, config *Config) error {
	logger := observationCtx.Logger

	// Load and validate configuration.
	if err := config.Validate(); err != nil {
		return errors.Wrap(err, "failed to validate configuration")
	}

	// Create a database connection.
	sqlDB, err := getDB(observationCtx)
	if err != nil {
		return errors.Wrap(err, "initializing database stores")
	}
	db := database.NewDB(observationCtx.Logger, sqlDB)

	// Initialize the keyring.
	err = keyring.Init(ctx)
	if err != nil {
		return errors.Wrap(err, "initializing keyring")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	routines := []goroutine.BackgroundRoutine{
		makeGRPCServer(logger, db, config),
		reconciler.New(
			ctx,
			db,
			config.ReconcilerInterval,
		),
	}

	// Register recorder in all routines that support it.
	recorderCache := recorder.GetCache()
	rec := recorder.New(observationCtx.Logger, env.MyName, recorderCache)
	for _, r := range routines {
		if recordable, ok := r.(recorder.Recordable); ok {
			recordable.SetJobName("tenantmanager")
			recordable.RegisterRecorder(rec)
			rec.Register(recordable)
		}
	}
	rec.RegistrationDone()

	debugserverEndpoints.tenantListEndpoint = func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode([]any{}); err != nil {
			logger.Error("failed to encode tenants", log.Error(err))
		}
	}

	logger.Info("tenantmanager: listening", log.String("addr", config.ListenAddress))

	// We're ready!
	ready()

	// Launch all routines!
	return goroutine.MonitorBackgroundRoutines(ctx, routines...)
}

// makeGRPCServer creates a new *grpc.Server for the gitserver endpoints and registers
// it with methods on the given server.
func makeGRPCServer(logger log.Logger, db dbutil.DB, c *Config) goroutine.BackgroundRoutine {
	grpcServer := defaults.NewServer(logger)
	proto.RegisterTenantManagerServiceServer(grpcServer, internal.NewTenantManagerServiceServer(logger, store.New(db), &internal.GRPCTenantManagerServiceConfig{
		ExhaustiveRequestLoggingEnabled: c.ExhaustiveRequestLoggingEnabled,
	}))

	return grpcserver.NewFromAddr(logger, c.ListenAddress, grpcServer)
}

// getDB initializes a connection to the database and returns a *sql.DB.
func getDB(observationCtx *observation.Context) (*sql.DB, error) {
	// Tenantmanager is an internal actor. We rely on the frontend to do authz checks
	// for user requests.
	//
	// This call to SetProviders is here so that calls to GetProviders don't block.
	authz.SetProviders(true, []authz.Provider{})

	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})
	return connections.EnsureNewFrontendDB(observationCtx, dsn, "tenantmanager")
}
