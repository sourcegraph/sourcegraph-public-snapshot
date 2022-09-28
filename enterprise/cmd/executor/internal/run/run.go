package run

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Masterminds/semver"
	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/config"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/ignite"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/janitor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/metrics"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func RunRun(cliCtx *cli.Context, logger log.Logger, cfg *config.Config) error {
	shouldRunVerify := cliCtx.Bool("verify")

	if shouldRunVerify {
		// TODO: validate docker is installed.
		// TODO: validate git is installed.
		// TODO: validate src-cli is installed and a good version, helper for that at the end of the file..
	}

	if err := cfg.Validate(); err != nil {
		logger.Error(err.Error())
		// logger.Error("failed to read config", log.Error(err))
		os.Exit(1)
	}

	// Initialize tracing/metrics
	observationContext := &observation.Context{
		Logger:     log.Scoped("service", "executor service"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	// Determine telemetry data.
	telemetryOptions := func() apiclient.TelemetryOptions {
		// Run for at most 5s to get telemetry options.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return apiclient.NewTelemetryOptions(ctx, cfg.UseFirecracker)
	}()
	logger.Info("Telemetry information gathered", log.String("info", fmt.Sprintf("%+v", telemetryOptions)))
	if cfg.UseFirecracker {
		want, err := semver.NewVersion(config.DefaultIgniteVersion)
		if err != nil {
			return err
		}
		have, err := semver.NewVersion(telemetryOptions.IgniteVersion)
		if err != nil {
			logger.Warn("failed to parse ignite version", log.Error(err))
		} else if !want.Equal(have) {
			logger.Warn("using unsupported ignite version, if things don't work alright, consider switching to the supported version", log.String("have", have.String()), log.String("want", want.String()))
		}
	}

	apiWorkerOptions := apiWorkerOptions(cfg, telemetryOptions)

	gatherer := metrics.MakeExecutorMetricsGatherer(log.Scoped("executor-worker.metrics-gatherer", ""), prometheus.DefaultGatherer, apiWorkerOptions.NodeExporterEndpoint, apiWorkerOptions.DockerRegistryNodeExporterEndpoint)
	queueStore := apiclient.New(apiWorkerOptions.ClientOptions, gatherer, observationContext)

	nameSet := janitor.NewNameSet()
	ctx, cancel := context.WithCancel(cliCtx.Context)
	worker := worker.NewWorker(nameSet, queueStore, apiWorkerOptions, observationContext)

	routines := []goroutine.BackgroundRoutine{
		worker,
	}

	if cfg.UseFirecracker {
		routines = append(routines, janitor.NewOrphanedVMJanitor(
			cfg.VMPrefix,
			nameSet,
			cfg.CleanupTaskInterval,
			janitor.NewMetrics(observationContext),
		))

		mustRegisterVMCountMetric(observationContext, cfg.VMPrefix)

		// If this causes harm, we can disable it.
		// if _, ok := os.LookupEnv("EXECUTOR_SKIP_FIRECRACKER_SETUP"); !ok {
		// 	if err := prepareFirecracker(ctx, logger, config.FirecrackerOptions()); err != nil {
		// 		logger.Error("failed to prepare firecracker environment", log.Error(err))
		// 		os.Exit(1)
		// 	}
		// }
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
	return nil
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
