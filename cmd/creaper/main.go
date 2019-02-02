package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/creaper/logger"

	"github.com/sourcegraph/sourcegraph/cmd/creaper/creaper"

	"code.cloudfoundry.org/bytefmt"
	"github.com/sourcegraph/sourcegraph/pkg/env"
)

var maxCacheSize = flag.String(
	"maxSize",
	"1g",
	"max cache size (examples: 1048576k, 1024m, 1g)",
)

var checkFrequency = flag.Duration(
	"frequency",
	time.Minute,
	"frequency with which the creaper should check disk usage",
)

var cacheDir = flag.String(
	"cacheDir",
	"/tmp",
	"cache directory to monitor",
)

var force = flag.Bool(
	"force",
	false,
	"turn off sanity checking",
)

func main() {
	env.Lock()
	env.HandleHelpFlag()
	flag.Parse()

	maxCacheSizeBytes, bytefmtErr := bytefmt.ToBytes(*maxCacheSize)
	if bytefmtErr != nil {
		flag.Usage()
		os.Exit(2)
	}

	// Unless force is specified, do some sanity checking.
	if !*force {
		if !strings.HasPrefix(*cacheDir, os.TempDir()) {
			fmt.Println("Specified cache directory is not a temporary folder. Use --force to override")
			os.Exit(3)
		}

	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	sigtermChan := make(chan os.Signal, 1)
	signal.Notify(sigtermChan, os.Interrupt, os.Kill)

	shutdownWaitGroup := &sync.WaitGroup{}

	shutdownWaitGroup.Add(1)
	go func() {
		creaper.Reap(
			logger.WithLogger(ctx, "cmd", "creaper"),
			*cacheDir,
			*checkFrequency,
			maxCacheSizeBytes)

		shutdownWaitGroup.Done()
	}()

	go func() {
		// Wait for shutdown signal.
		_ = <-sigtermChan

		logger.Info(ctx, "Request to shutdown received.")

		cancelFunc()
	}()

	shutdownWaitGroup.Wait()

	logger.Info(ctx, "Shutdown complete")
}
