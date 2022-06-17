package shared

import (
	"context"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	sentrylib "github.com/getsentry/sentry-go"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/fetcher"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/gitserver"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/api"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/profiler"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

const addr = ":3184"

type SetupFunc func(observationContext *observation.Context, gitserverClient gitserver.GitserverClient, repositoryFetcher fetcher.RepositoryFetcher) (types.SearchFunc, func(http.ResponseWriter, *http.Request), []goroutine.BackgroundRoutine, string, error)

func Main(setup SetupFunc) {
	// Initialization
	env.HandleHelpFlag()
	conf.Init()
	logging.Init()
	liblog := log.Init(log.Resource{
		Name:       env.MyName,
		Version:    version.Version(),
		InstanceID: hostname.Get(),
	}, log.NewSentrySinkWithOptions(sentrylib.ClientOptions{SampleRate: 0.2})) // Experimental: DevX is observing how sampling affects the errors signal

	defer liblog.Sync()
	go conf.Watch(liblog.Update(conf.GetLogSinks))
	tracer.Init(conf.DefaultClient())
	trace.Init()
	profiler.Init()

	routines := []goroutine.BackgroundRoutine{}

	// Initialize tracing/metrics
	logger := log.Scoped("service", "the symbols service")
	observationContext := &observation.Context{
		Logger:     logger,
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
		HoneyDataset: &honey.Dataset{
			Name:       "codeintel-symbols",
			SampleRate: 5,
		},
	}

	// Run setup
	gitserverClient := gitserver.NewClient(observationContext)
	repositoryFetcherConfig := types.LoadRepositoryFetcherConfig(env.BaseConfig{})
	repositoryFetcher := fetcher.NewRepositoryFetcher(gitserverClient, repositoryFetcherConfig.MaxTotalPathsLength, int64(repositoryFetcherConfig.MaxFileSizeKb)*1000, observationContext)
	searchFunc, handleStatus, newRoutines, ctagsBinary, err := setup(observationContext, gitserverClient, repositoryFetcher)
	if err != nil {
		logger.Fatal("Failed to set up", log.Error(err))
	}
	routines = append(routines, newRoutines...)

	// Start debug server
	ready := make(chan struct{})
	go debugserver.NewServerRoutine(ready).Start()

	// Create HTTP server
	server := httpserver.NewFromAddr(addr, &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Handler:      actor.HTTPMiddleware(ot.HTTPMiddleware(trace.HTTPMiddleware(logger, api.NewHandler(searchFunc, handleStatus, ctagsBinary), conf.DefaultClient()))),
	})
	routines = append(routines, server)

	// Mark health server as ready and go!
	close(ready)
	goroutine.MonitorBackgroundRoutines(context.Background(), routines...)
}
