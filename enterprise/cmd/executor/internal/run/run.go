package run

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient/queue"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/config"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/ignite"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/janitor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func RunRun(cliCtx *cli.Context, logger log.Logger, cfg *config.Config) error {
	return StandaloneRunRun(cliCtx.Context, logger, cfg, cliCtx.Bool("verify"))
}

func StandaloneRunRun(ctx context.Context, logger log.Logger, cfg *config.Config, runVerifyChecks bool) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	logger = log.Scoped("service", "executor service")

	// Initialize tracing/metrics
	observationCtx := observation.NewContext(logger)

	// Determine telemetry data.
	queueTelemetryOptions := func() queue.TelemetryOptions {
		// Run for at most 5s to get telemetry options.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return newQueueTelemetryOptions(ctx, cfg.UseFirecracker, logger)
	}()
	logger.Debug("Telemetry information gathered", log.String("info", fmt.Sprintf("%+v", queueTelemetryOptions)))

	opts := apiWorkerOptions(cfg, queueTelemetryOptions)

	// TODO: This is too similar to the RunValidate func. Make it share even more code.
	if runVerifyChecks {
		// Then, validate all tools that are required are installed.
		if err := validateToolsRequired(cfg.UseFirecracker); err != nil {
			return err
		}

		// Validate git is of the right version.
		if err := validateGitVersion(ctx); err != nil {
			return err
		}

		// TODO: Validate access token.
		// Validate src-cli is of a good version, rely on the connected instance to tell
		// us what "good" means.
		client, err := apiclient.NewBaseClient(opts.QueueOptions.BaseClientOptions)
		if err != nil {
			return err
		}
		if err := validateSrcCLIVersion(ctx, client, opts.QueueOptions.BaseClientOptions.EndpointOptions); err != nil {
			return err
		}

		if cfg.UseFirecracker {
			// Validate ignite is installed.
			if err := validateIgniteInstalled(ctx); err != nil {
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
	ctx, cancel := context.WithCancel(ctx)
	wrk, err := worker.NewWorker(observationCtx, nameSet, opts)
	if err != nil {
		cancel()
		return err
	}

	routines := []goroutine.BackgroundRoutine{wrk}

	if cfg.UseFirecracker {
		routines = append(routines, janitor.NewOrphanedVMJanitor(
			cfg.VMPrefix,
			nameSet,
			cfg.CleanupTaskInterval,
			janitor.NewMetrics(observationCtx),
		))

		mustRegisterVMCountMetric(logger, observationCtx, cfg.VMPrefix)
	}

	go func() {
		// Block until the worker has exited. The executor worker is unique
		// in that we want a maximum runtime and/or number of jobs to be
		// executed by a single instance, after which the service should shut
		// down without error.
		wrk.Wait()

		// Once the worker has finished its current set of jobs and stops
		// the dequeue loop, we want to finish off the rest of the sibling
		// routines so that the service can shut down.
		cancel()
	}()

	goroutine.MonitorBackgroundRoutines(ctx, routines...)
	return nil
}

func mustRegisterVMCountMetric(logger log.Logger, observationCtx *observation.Context, prefix string) {
	observationCtx.Registerer.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_executor_vms_total",
		Help: "Total number of running VMs.",
	}, func() float64 {
		runningVMsByName, err := ignite.ActiveVMsByName(context.Background(), prefix, false)
		if err != nil {
			logger.Error("Failed to determine number of running VMs", log.Error(err))
		}

		return float64(len(runningVMsByName))
	}))
}
