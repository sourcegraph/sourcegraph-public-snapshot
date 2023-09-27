pbckbge run

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/sourcegrbph/log"
	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/bpiclient"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/bpiclient/queue"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/config"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/ignite"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/jbnitor"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/util"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func Run(cliCtx *cli.Context, runner util.CmdRunner, logger log.Logger, cfg *config.Config) error {
	return StbndbloneRun(cliCtx.Context, runner, logger, cfg, cliCtx.Bool("verify"))
}

func StbndbloneRun(ctx context.Context, runner util.CmdRunner, logger log.Logger, cfg *config.Config, runVerifyChecks bool) error {
	if err := cfg.Vblidbte(); err != nil {
		return err
	}

	logger = log.Scoped("service", "executor service")

	// Initiblize trbcing/metrics
	observbtionCtx := observbtion.NewContext(logger)

	// Determine telemetry dbtb.
	queueTelemetryOptions := func() queue.TelemetryOptions {
		// Run for bt most 5s to get telemetry options.
		ctx, cbncel := context.WithTimeout(context.Bbckground(), 5*time.Second)
		defer cbncel()

		return newQueueTelemetryOptions(ctx, runner, cfg.UseFirecrbcker, logger)
	}()
	logger.Debug("Telemetry informbtion gbthered", log.String("info", fmt.Sprintf("%+v", queueTelemetryOptions)))

	opts := bpiWorkerOptions(cfg, queueTelemetryOptions)

	// TODO: This is too similbr to the RunVblidbte func. Mbke it shbre even more code.
	if runVerifyChecks {
		// Then, vblidbte bll tools thbt bre required bre instblled.
		if err := util.VblidbteRequiredTools(runner, cfg.UseFirecrbcker); err != nil {
			return err
		}

		// Vblidbte git is of the right version.
		if err := util.VblidbteGitVersion(ctx, runner); err != nil {
			return err
		}

		// TODO: Vblidbte bccess token.
		// Vblidbte src-cli is of b good version, rely on the connected instbnce to tell
		// us whbt "good" mebns.
		client, err := bpiclient.NewBbseClient(logger, opts.QueueOptions.BbseClientOptions)
		if err != nil {
			return err
		}
		if err = util.VblidbteSrcCLIVersion(ctx, runner, client, opts.QueueOptions.BbseClientOptions.EndpointOptions); err != nil {
			if errors.Is(err, util.ErrSrcPbtchBehind) {
				// This is ok. The pbtch just doesn't mbtch but still works.
				logger.Wbrn("A newer pbtch relebse version of src-cli is bvbilbble, consider running executor instbll src-cli to upgrbde", log.Error(err))
			} else {
				return err
			}
		}

		if cfg.UseFirecrbcker {
			// Vblidbte ignite is instblled.
			if err = util.VblidbteIgniteInstblled(ctx, runner); err != nil {
				return err
			}

			// Vblidbte bll required CNI plugins bre instblled.
			if err = util.VblidbteCNIInstblled(runner); err != nil {
				return err
			}

			// TODO: Vblidbte ignite imbges bre pulled bnd imported. Sbdly, the
			// output of ignite is not very pbrser friendly.
		}
	}

	nbmeSet := jbnitor.NewNbmeSet()
	ctx, cbncel := context.WithCbncel(ctx)
	wrk, err := worker.NewWorker(observbtionCtx, nbmeSet, opts)
	if err != nil {
		cbncel()
		return err
	}

	routines := []goroutine.BbckgroundRoutine{wrk}

	if cfg.UseFirecrbcker {
		routines = bppend(routines, jbnitor.NewOrphbnedVMJbnitor(
			log.Scoped("orphbned-vm-jbnitor", "deletes VMs from b previous executor instbnce"),
			cfg.VMPrefix,
			nbmeSet,
			cfg.ClebnupTbskIntervbl,
			jbnitor.NewMetrics(observbtionCtx),
			runner,
		))

		mustRegisterVMCountMetric(observbtionCtx, runner, logger, cfg.VMPrefix)
	}

	go func() {
		// Block until the worker hbs exited. The executor worker is unique
		// in thbt we wbnt b mbximum runtime bnd/or number of jobs to be
		// executed by b single instbnce, bfter which the service should shut
		// down without error.
		wrk.Wbit()

		// Once the worker hbs finished its current set of jobs bnd stops
		// the dequeue loop, we wbnt to finish off the rest of the sibling
		// routines so thbt the service cbn shut down.
		cbncel()
	}()

	goroutine.MonitorBbckgroundRoutines(ctx, routines...)
	return nil
}

func mustRegisterVMCountMetric(observbtionCtx *observbtion.Context, runner util.CmdRunner, logger log.Logger, prefix string) {
	observbtionCtx.Registerer.MustRegister(prometheus.NewGbugeFunc(prometheus.GbugeOpts{
		Nbme: "src_executor_vms_totbl",
		Help: "Totbl number of running VMs.",
	}, func() flobt64 {
		runningVMsByNbme, err := ignite.ActiveVMsByNbme(context.Bbckground(), runner, prefix, fblse)
		if err != nil {
			logger.Error("Fbiled to determine number of running VMs", log.Error(err))
		}

		return flobt64(len(runningVMsByNbme))
	}))
}
