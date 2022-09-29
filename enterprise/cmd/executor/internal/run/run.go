package run

import (
	"context"
	"fmt"
	"time"

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
	if err := cfg.Validate(); err != nil {
		return err
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

		return newTelemetryOptions(ctx, cfg.UseFirecracker)
	}()
	logger.Info("Telemetry information gathered", log.String("info", fmt.Sprintf("%+v", telemetryOptions)))

	apiWorkerOptions := apiWorkerOptions(cfg, telemetryOptions)

	gatherer := metrics.MakeExecutorMetricsGatherer(log.Scoped("executor-worker.metrics-gatherer", ""), prometheus.DefaultGatherer, apiWorkerOptions.NodeExporterEndpoint, apiWorkerOptions.DockerRegistryNodeExporterEndpoint)
	queueStore := apiclient.New(apiWorkerOptions.ClientOptions, gatherer, observationContext)

	// TODO: This is too similar to the RunValidate func. Make it share even more code.
	if cliCtx.Bool("verify") {
		// Then, validate all tools that are required are installed.
		if err := validateToolsRequired(cfg.UseFirecracker); err != nil {
			return err
		}

		// Validate git is of the right version.
		if err := validateGitVersion(telemetryOptions); err != nil {
			return err
		}

		// TODO: Validate access token.
		// Validate src-cli is of a good version, rely on the connected instance to tell
		// us what "good" means.
		if err := validateSrcCLIVersion(cliCtx.Context, logger, queueStore); err != nil {
			return err
		}

		if cfg.UseFirecracker {
			// Validate ignite is installed.
			if err := validateIgniteInstalled(cliCtx.Context); err != nil {
				return err
			}

			// Validate all required CNI plugins are installed.
			if err := validateCNIInstalled(); err != nil {
				return err
			}

			// TODO: Validate ignite images are pulled and imported. Sadly, the
			// output of ignite is not very parser friendly.
		}
	}

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
