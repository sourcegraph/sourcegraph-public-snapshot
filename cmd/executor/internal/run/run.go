package run

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/apiclient/queue"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/config"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/ignite"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/janitor"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/util"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func Run(cliCtx *cli.Context, runner util.CmdRunner, logger log.Logger, cfg *config.Config) error {
	return StandaloneRun(cliCtx.Context, runner, logger, cfg, cliCtx.Bool("verify"))
}

func StandaloneRun(ctx context.Context, runner util.CmdRunner, logger log.Logger, cfg *config.Config, runVerifyChecks bool) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	logger = log.Scoped("service")

	// Initialize tracing/metrics
	observationCtx := observation.NewContext(logger)

	// Determine telemetry data.
	queueTelemetryOptions := func() queue.TelemetryOptions {
		// Run for at most 5s to get telemetry options.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return newQueueTelemetryOptions(ctx, runner, cfg.UseFirecracker, logger)
	}()
	logger.Debug("Telemetry information gathered", log.String("info", fmt.Sprintf("%+v", queueTelemetryOptions)))

	opts := apiWorkerOptions(cfg, queueTelemetryOptions)

	// TODO: This is too similar to the RunValidate func. Make it share even more code.
	if runVerifyChecks {
		// Then, validate all tools that are required are installed.
		if err := util.ValidateRequiredTools(runner, cfg.UseFirecracker); err != nil {
			return err
		}

		// Validate git is of the right version.
		if err := util.ValidateGitVersion(ctx, runner); err != nil {
			return err
		}

		// TODO: Validate access token.
		// Validate src-cli is of a good version, rely on the connected instance to tell
		// us what "good" means.
		client, err := apiclient.NewBaseClient(logger, opts.QueueOptions.BaseClientOptions)
		if err != nil {
			return err
		}
		if err = util.ValidateSrcCLIVersion(ctx, runner, client); err != nil {
			if errors.Is(err, util.ErrSrcPatchBehind) {
				// This is ok. The patch just doesn't match but still works.
				logger.Warn("A newer patch release version of src-cli is available, consider running executor install src-cli to upgrade", log.Error(err))
			} else {
				return err
			}
		}

		if cfg.UseFirecracker {
			// Validate ignite is installed.
			if err = util.ValidateIgniteInstalled(ctx, runner); err != nil {
				return err
			}

			// Validate all required CNI plugins are installed.
			if err = util.ValidateCNIInstalled(runner); err != nil {
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
			log.Scoped("orphaned-vm-janitor"),
			cfg.VMPrefix,
			nameSet,
			cfg.CleanupTaskInterval,
			janitor.NewMetrics(observationCtx),
			runner,
		))

		mustRegisterVMCountMetric(observationCtx, runner, logger, cfg.VMPrefix)
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

func mustRegisterVMCountMetric(observationCtx *observation.Context, runner util.CmdRunner, logger log.Logger, prefix string) {
	observationCtx.Registerer.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_executor_vms_total",
		Help: "Total number of running VMs.",
	}, func() float64 {
		runningVMsByName, err := ignite.ActiveVMsByName(context.Background(), runner, prefix, false)
		if err != nil {
			logger.Error("Failed to determine number of running VMs", log.Error(err))
		}

		return float64(len(runningVMsByName))
	}))
}
