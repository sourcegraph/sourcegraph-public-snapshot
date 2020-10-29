package main

import (
	"log"

	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/metrics"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/worker"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/sqliteutil"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
)

func main() {
	config := &Config{}
	config.Load()

	env.Lock()
	env.HandleHelpFlag()
	logging.Init()
	tracer.Init()
	trace.Init(true)

	if err := config.Validate(); err != nil {
		log.Fatalf("failed to load config: %s", err)
	}

	sqliteutil.MustRegisterSqlite3WithPcre()

	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	store := store.NewObserved(mustInitializeStore(), observationContext)
	codeIntelDB := mustInitializeCodeIntelDatabase()
	MustRegisterQueueMonitor(observationContext.Registerer, store)

	goroutine.MonitorBackgroundRoutines(
		goroutine.NoopStop(debugserver.NewServerRoutine()),
		httpserver.New(Port, func(router *mux.Router) {}),
		worker.NewWorker(
			store,
			codeIntelDB,
			bundles.New(codeIntelDB, observationContext, config.BundleManagerURL),
			gitserver.DefaultClient,
			config.WorkerPollInterval,
			config.WorkerConcurrency,
			config.WorkerBudget,
			metrics.NewWorkerMetrics(observationContext),
			observationContext,
		),
	)
}
