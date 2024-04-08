package shared

import (
	"context"

	"net/http"
	"time"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/syntactic_indexing/jobstore"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
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

	jobStore, err := jobstore.NewStore(observationCtx, "syntactic-code-intel-indexer")
	if err != nil {
		return errors.Wrap(err, "initializing worker store")
	}

	indexingWorker := NewIndexingWorker(ctx, observationCtx, jobStore, *config.IndexingWorkerConfig)

	// Initialize health server
	server := httpserver.NewFromAddr(config.ListenAddress, &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Handler:      httpserver.NewHandler(nil),
	})

	// Go!
	goroutine.MonitorBackgroundRoutines(ctx, server, indexingWorker)

	return nil
}
