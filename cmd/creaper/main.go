package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
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

	ctx, _ := context.WithCancel(context.Background())

	creaper.Reap(
		logger.WithLogger(ctx, "cmd", "creaper"),
		*cacheDir,
		*checkFrequency,
		maxCacheSizeBytes)
}
