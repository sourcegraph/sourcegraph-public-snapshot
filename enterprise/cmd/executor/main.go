package main

import (
	"context"
	"log"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/ignite"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/janitor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

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

	// Initialize tracing/metrics
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	// Ready immediately
	ready := make(chan struct{})
	close(ready)
	go debugserver.NewServerRoutine(ready).Start()

	nameSet := janitor.NewNameSet()

	routines := []goroutine.BackgroundRoutine{
		worker.NewWorker(nameSet, config.APIWorkerOptions(), observationContext),
	}
	if config.UseFirecracker {
		routines = append(routines, janitor.NewOrphanedVMJanitor(
			config.VMPrefix,
			nameSet,
			config.CleanupTaskInterval,
			janitor.NewMetrics(observationContext),
		))

		mustRegisterVMCountMetric(observationContext, config.VMPrefix)
	}
	goroutine.MonitorBackgroundRoutines(context.Background(), routines...)
}

func makeWorkerMetrics(queueName string) workerutil.WorkerMetrics {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	return workerutil.NewMetrics(observationContext, "executor_processor", map[string]string{
		"queue": queueName,
	})
}

func mustRegisterVMCountMetric(observationContext *observation.Context, prefix string) {
	observationContext.Registerer.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_executor_vms_total",
		Help: "Total number of running VMs.",
	}, func() float64 {
		runningVMsByName, err := ignite.ActiveVMsByName(context.Background(), prefix, false)
		if err != nil {
			log15.Error("Failed to determine number of running VMs", "error", err)
		}

		return float64(len(runningVMsByName))
	}))
}
