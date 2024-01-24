package shared

import (
	"context"

	"net/http"
	"time"

	"github.com/sourcegraph/log"
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
		log.String("path to scip-treesitter CLI", config.CliPath),
		log.String("API address", config.ListenAddress))

	// Initialize health server
	server := httpserver.NewFromAddr(config.ListenAddress, &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Handler:      httpserver.NewHandler(nil),
	})

	// Go!
	goroutine.MonitorBackgroundRoutines(ctx, server)

	return nil
}
