package shared

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/sourcegraph/log"

	codeintelshared "github.com/sourcegraph/sourcegraph/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/syntactic_indexing/jobstore"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func Main(ctx context.Context, observationCtx *observation.Context, ready service.ReadyFunc, config Config) error {
	logger := observationCtx.Logger

	if err := keyring.Init(ctx); err != nil {
		return errors.Wrap(err, "initializing keyring")
	}

	logger.Info("Syntactic code intel worker running",
		log.String("path to scip-syntax CLI", config.IndexingWorkerConfig.CliPath),
		log.String("API address", config.ListenAddress))

	name := "syntactic-codeintel-worker"

	db, err := initDB(observationCtx, name)
	if err != nil {
		return errors.Wrap(err, "initializing frontend db")
	}

	jobStore, err := jobstore.NewStoreWithDB(observationCtx, db)
	if err != nil {
		return errors.Wrap(err, "initializing worker store")
	}

	codeintelSqlDB, err := initCodeintelDB(observationCtx, name)
	if err != nil {
		return errors.Wrap(err, "initializing codeintel db")
	}
	codeintelDB := codeintelshared.NewCodeIntelDB(logger, codeintelSqlDB)

	indexingWorker, err := NewIndexingWorker(ctx,
		observationCtx,
		jobStore,
		*config.IndexingWorkerConfig,
		db,
		codeintelDB,
		gitserver.NewClient(name),
	)

	if err != nil {
		return errors.Wrap(err, "creating syntactic codeintel indexing worker")
	}

	// Initialize health server
	server := httpserver.NewFromAddr(config.ListenAddress, &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Handler:      httpserver.NewHandler(nil),
	})

	// Go!
	return goroutine.MonitorBackgroundRoutines(ctx, server, indexingWorker)
}

func initCodeintelDB(observationCtx *observation.Context, name string) (*sql.DB, error) {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.CodeIntelPostgresDSN
	})

	sqlDB, err := connections.EnsureNewCodeIntelDB(observationCtx, dsn, name)

	if err != nil {
		log.Scoped("init db ("+name+")").Fatal("Failed to connect to codeintel database", log.Error(err))
		return nil, err
	}

	return sqlDB, nil
}

func initDB(observationCtx *observation.Context, name string) (database.DB, error) {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})

	sqlDB, err := connections.EnsureNewFrontendDB(observationCtx, dsn, name)

	if err != nil {
		log.Scoped("init db ("+name+")").Fatal("Failed to connect to frontend database", log.Error(err))
		return nil, err
	}

	return database.NewDB(observationCtx.Logger, sqlDB), nil
}
