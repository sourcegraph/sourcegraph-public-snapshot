package shared

import (
	"context"
	"fmt"

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

const addr = ":3188"

func Main(ctx context.Context, observationCtx *observation.Context, ready service.ReadyFunc, config Config) error {
	logger := observationCtx.Logger

	// Initialize tracing/metrics
	// observationCtx = observation.NewContext(logger, observation.Honeycomb(&honey.Dataset{
	// 	Name: "syntactic-code-intel-worker",
	// }))

	if err := keyring.Init(ctx); err != nil {
		return errors.Wrap(err, "initializing keyring")
	}

	logger.Info("Syntactic code intel worker running", log.String("path to scip-treesitter CLI", config.CliPath))

	fmt.Println(config.CliPath)

	// // Initialize health server
	server := httpserver.NewFromAddr(addr, &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Handler:      httpserver.NewHandler(nil),
	})

	// // Go!
	goroutine.MonitorBackgroundRoutines(ctx, server)

	return nil
}
