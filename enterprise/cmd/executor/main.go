package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/ignite"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/janitor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"

	// This import is required to force a binary hash change when the src-cli version is bumped.
	_ "github.com/sourcegraph/sourcegraph/internal/src-cli"
)

func main() {
	config := &Config{}
	config.Load()

	env.Lock()
	env.HandleHelpFlag()

	logging.Init()
	liblog := log.Init(log.Resource{
		Name:       env.MyName,
		Version:    version.Version(),
		InstanceID: hostname.Get(),
	})
	defer liblog.Sync()
	trace.Init()

	logger := log.Scoped("executor", "the executor service polls the public frontend API for work to perform")

	if err := config.Validate(); err != nil {
		logger.Error("failed to read config", log.Error(err))
		os.Exit(1)
	}

	// Initialize tracing/metrics
	observationContext := &observation.Context{
		Logger:     log.Scoped("service", "executor service"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	// Ready immediately
	ready := make(chan struct{})
	close(ready)
	go debugserver.NewServerRoutine(ready).Start()

	// Determine telemetry data.
	telemetryOptions := func() apiclient.TelemetryOptions {
		// Run for at most 5s to get telemetry options.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return apiclient.NewTelemetryOptions(ctx)
	}()
	logger.Info("Telemetry information gathered", log.String("info", fmt.Sprintf("%+v", telemetryOptions)))

	nameSet := janitor.NewNameSet()
	ctx, cancel := context.WithCancel(context.Background())
	worker := worker.NewWorker(nameSet, config.APIWorkerOptions(telemetryOptions), observationContext)

	routines := []goroutine.BackgroundRoutine{
		worker,
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

	go func() {
		// Block until the worker has exited. The executor worker is unique
		// in that we want a maximum runtime and/or number of jobs to be
		// executed by a single instance, after which the service should shut
		// down without error.
		worker.Wait()

		// Once the worker has finished its current set of jobs and stops
		// the dequeue loop, we want to finish off the rest of the sibling
		// routines so that the service can shut down.
		cancel()
	}()

	goroutine.MonitorBackgroundRoutines(ctx, routines...)
}

func makeWorkerMetrics(queueName string) workerutil.WorkerMetrics {
	observationContext := &observation.Context{
		Logger:     log.Scoped("executor_processor", "executor worker processor"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	return workerutil.NewMetrics(observationContext, "executor_processor",
		// derived from historic data, ideally we will use spare high-res histograms once they're a reality
		// 										 30s 1m	 2.5m 5m   7.5m 10m  15m  20m	30m	  45m	1hr
		workerutil.WithDurationBuckets([]float64{30, 60, 150, 300, 450, 600, 900, 1200, 1800, 2700, 3600}),
		workerutil.WithLabels(map[string]string{
			"queue": queueName,
		}),
	)
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
