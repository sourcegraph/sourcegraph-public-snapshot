package main

import (
	"context"
	"log"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/metrics"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/worker"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	uploadstore "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/upload_store"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
)

const addr = ":3188"

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

	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	store := store.NewObserved(mustInitializeStore(), observationContext)
	codeIntelDB := mustInitializeCodeIntelDatabase()
	MustRegisterQueueMonitor(observationContext.Registerer, store)

	server, err := httpserver.NewFromAddr(addr, httpserver.NewHandler(nil), httpserver.Options{})
	if err != nil {
		log.Fatalf("Failed to create listener: %s", err)
	}

	uploadStore, err := uploadstore.Create(context.Background(), config.UploadStoreConfig)
	if err != nil {
		log.Fatalf("failed to initialize upload store: %s", err)
	}

	debugServer, err := debugserver.NewServerRoutine()
	if err != nil {
		log.Fatalf("Failed to create listener: %s", err)
	}
	go debugServer.Start()

	goroutine.MonitorBackgroundRoutines(
		context.Background(),
		server,
		worker.NewWorker(
			store,
			codeIntelDB,
			uploadStore,
			gitserver.DefaultClient,
			config.WorkerPollInterval,
			config.WorkerConcurrency,
			config.WorkerBudget,
			metrics.NewWorkerMetrics(observationContext),
			observationContext,
		),
	)
}
