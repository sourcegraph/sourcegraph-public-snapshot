package main

import (
	"context"
	"log"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/apiworker"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

const addr = ":3192"

func main() {
	config := &Config{}
	config.Load()

	env.Lock()
	env.HandleHelpFlag()

	logging.Init()
	trace.Init(false)

	if err := config.Validate(); err != nil {
		log.Fatalf("failed to read config: %s", err)
	}

	debugServer, err := debugserver.NewServerRoutine()
	if err != nil {
		log.Fatalf("Failed to create listener: %s", err)
	}
	go debugServer.Start()

	routines := []goroutine.BackgroundRoutine{
		apiworker.NewWorker(config.APIWorkerOptions(nil)),
	}
	if !config.DisableHealthServer {
		server, err := httpserver.NewFromAddr(addr, httpserver.NewHandler(nil), httpserver.Options{})
		if err != nil {
			log.Fatalf("Failed to create listener: %s", err)
		}

		routines = append(routines, server)
	}

	goroutine.MonitorBackgroundRoutines(context.Background(), routines...)
}

func makeWorkerMetrics() workerutil.WorkerMetrics {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		"executor_queue_processor",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of records processed"),
	)

	return workerutil.WorkerMetrics{
		HandleOperation: observationContext.Operation(observation.Op{
			Name:         "Processor.Process",
			MetricLabels: []string{"process"},
			Metrics:      metrics,
		}),
	}
}
