package shared

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/api"
	sqlite "github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database"
	sharedtypes "github.com/sourcegraph/sourcegraph/cmd/symbols/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/profiler"
	"github.com/sourcegraph/sourcegraph/internal/sentry"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
)

const addr = ":3184"

func Main(setup func(config *Config, observationContext *observation.Context) (sharedtypes.SearchFunc, []goroutine.BackgroundRoutine)) {
	config.Load()

	// Set up Google Cloud Profiler when running in Cloud
	if err := profiler.Init(); err != nil {
		log.Fatalf("Failed to start profiler: %v", err)
	}

	env.Lock()
	env.HandleHelpFlag()
	conf.Init()
	logging.Init()
	tracer.Init(conf.DefaultClient())
	sentry.Init(conf.DefaultClient())
	trace.Init()

	if err := config.Validate(); err != nil {
		log.Fatalf("Failed to load configuration: %s", err)
	}

	// Ensure we register our database driver before calling
	// anything that tries to open a SQLite database.
	sqlite.Init()

	if config.SanityCheck {
		fmt.Print("Running sanity check...")
		if err := sqlite.SanityCheck(); err != nil {
			fmt.Println("failed ❌", err)
			os.Exit(1)
		}

		fmt.Println("passed ✅")
		os.Exit(0)
	}

	// Initialize tracing/metrics
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
		HoneyDataset: &honey.Dataset{
			Name:       "codeintel-symbols",
			SampleRate: 5,
		},
	}

	// Start debug server
	ready := make(chan struct{})
	go debugserver.NewServerRoutine(ready).Start()

	searchFunc, routines := setup(config, observationContext)
	if searchFunc == nil {
		searchFunc, routines = SetupSqlite(config, observationContext)
	}

	apiHandler := api.NewHandler(searchFunc, config.Ctags.Command)

	server := httpserver.NewFromAddr(addr, &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Handler:      actor.HTTPMiddleware(ot.HTTPMiddleware(trace.HTTPMiddleware(apiHandler, conf.DefaultClient()))),
	})

	// Mark health server as ready and go!
	close(ready)
	goroutine.MonitorBackgroundRoutines(context.Background(), append([]goroutine.BackgroundRoutine{server}, routines...)...)
}
