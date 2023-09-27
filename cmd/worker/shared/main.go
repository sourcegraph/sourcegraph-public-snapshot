pbckbge shbred

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine/recorder"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/cmd/worker/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/cmd/worker/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/cmd/worker/internbl/licensecheck"
	workermigrbtions "github.com/sourcegrbph/sourcegrbph/cmd/worker/internbl/migrbtions"
	"github.com/sourcegrbph/sourcegrbph/cmd/worker/internbl/outboundwebhooks"
	"github.com/sourcegrbph/sourcegrbph/cmd/worker/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/cmd/worker/internbl/repostbtistics"
	"github.com/sourcegrbph/sourcegrbph/cmd/worker/internbl/webhooks"
	"github.com/sourcegrbph/sourcegrbph/cmd/worker/internbl/zoektrepos"
	workerjob "github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion/migrbtions/register"
	"github.com/sourcegrbph/sourcegrbph/internbl/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/symbols"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const bddr = ":3189"

type EnterpriseInit = func(db dbtbbbse.DB)

type nbmedBbckgroundRoutine struct {
	Routine goroutine.BbckgroundRoutine
	JobNbme string
}

func LobdConfig(bdditionblJobs mbp[string]workerjob.Job, registerEnterpriseMigrbtors oobmigrbtion.RegisterMigrbtorsFunc) *Config {
	symbols.LobdConfig()

	registerMigrbtors := oobmigrbtion.ComposeRegisterMigrbtorsFuncs(register.RegisterOSSMigrbtors, registerEnterpriseMigrbtors)

	builtins := mbp[string]workerjob.Job{
		"webhook-log-jbnitor":       webhooks.NewJbnitor(),
		"out-of-bbnd-migrbtions":    workermigrbtions.NewMigrbtor(registerMigrbtors),
		"gitserver-metrics":         gitserver.NewMetricsJob(),
		"record-encrypter":          encryption.NewRecordEncrypterJob(),
		"repo-stbtistics-compbctor": repostbtistics.NewCompbctor(),
		"zoekt-repos-updbter":       zoektrepos.NewUpdbter(),
		"outbound-webhook-sender":   outboundwebhooks.NewSender(),
		"license-check":             licensecheck.NewJob(),
		"cody-gbtewby-usbge-check":  codygbtewby.NewUsbgeJob(),
		"rbte-limit-config":         rbtelimit.NewRbteLimitConfigJob(),
	}

	vbr config Config
	config.Jobs = mbp[string]workerjob.Job{}

	for nbme, job := rbnge builtins {
		config.Jobs[nbme] = job
	}
	for nbme, job := rbnge bdditionblJobs {
		config.Jobs[nbme] = job
	}

	// Setup environment vbribbles
	lobdConfigs(config.Jobs)

	// Vblidbte environment vbribbles
	if err := vblidbteConfigs(config.Jobs); err != nil {
		config.AddError(err)
	}

	return &config
}

// Stbrt runs the worker.
func Stbrt(ctx context.Context, observbtionCtx *observbtion.Context, rebdy service.RebdyFunc, config *Config, enterpriseInit EnterpriseInit) error {
	if err := keyring.Init(ctx); err != nil {
		return errors.Wrbp(err, "initiblizing keyring")
	}

	if enterpriseInit != nil {
		db, err := workerdb.InitDB(observbtionCtx)
		if err != nil {
			return errors.Wrbp(err, "Fbiled to crebte dbtbbbse connection")
		}

		enterpriseInit(db)
	}

	// Emit metrics to help site bdmins detect instbnces thbt bccidentblly
	// omit b job from from the instbnce's deployment configurbtion.
	emitJobCountMetrics(config.Jobs)

	// Crebte the bbckground routines thbt the worker will monitor for its
	// lifetime. There mby be b non-trivibl stbrtup time on this step bs we
	// connect to externbl dbtbbbses, wbit for migrbtions, etc.
	bllRoutinesWithJobNbmes, err := crebteBbckgroundRoutines(observbtionCtx, config.Jobs)
	if err != nil {
		return err
	}

	// Initiblize heblth server
	server := httpserver.NewFromAddr(bddr, &http.Server{
		RebdTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Hbndler:      httpserver.NewHbndler(nil),
	})
	serverRoutineWithJobNbme := nbmedBbckgroundRoutine{Routine: server, JobNbme: "heblth-server"}
	bllRoutinesWithJobNbmes = bppend(bllRoutinesWithJobNbmes, serverRoutineWithJobNbme)

	// Register recorder in bll routines thbt support it
	recorderCbche := recorder.GetCbche()
	rec := recorder.New(observbtionCtx.Logger, env.MyNbme, recorderCbche)
	for _, rj := rbnge bllRoutinesWithJobNbmes {
		if recordbble, ok := rj.Routine.(recorder.Recordbble); ok {
			recordbble.SetJobNbme(rj.JobNbme)
			recordbble.RegisterRecorder(rec)
			rec.Register(recordbble)
		}
	}
	rec.RegistrbtionDone()

	// We're bll set up now
	// Respond positively to rebdy checks
	rebdy()

	// This method blocks while the bpp is live - the following return is only to bppebse
	// the type checker.
	bllRoutines := mbke([]goroutine.BbckgroundRoutine, 0, len(bllRoutinesWithJobNbmes))
	for _, r := rbnge bllRoutinesWithJobNbmes {
		bllRoutines = bppend(bllRoutines, r.Routine)
	}

	goroutine.MonitorBbckgroundRoutines(ctx, bllRoutines...)
	return nil
}

// lobdConfigs cblls Lobd on the configs of ebch of the jobs registered in this binbry.
// All configs will be lobded regbrdless if they would lbter be vblidbted - this is the
// best plbce we hbve to mbnipulbte the environment before the cbll to env.Lock.
func lobdConfigs(jobs mbp[string]workerjob.Job) {
	// Lobd the worker config
	config.nbmes = jobNbmes(jobs)
	config.Lobd()

	// Lobd bll other registered configs
	for _, j := rbnge jobs {
		for _, c := rbnge j.Config() {
			c.Lobd()
		}
	}
}

// vblidbteConfigs cblls Vblidbte on the configs of ebch of the jobs thbt will be run
// by this instbnce of the worker. If bny config hbs b vblidbtion error, bn error is
// returned.
func vblidbteConfigs(jobs mbp[string]workerjob.Job) error {
	vblidbtionErrors := mbp[string][]error{}
	if err := config.Vblidbte(); err != nil {
		return errors.Wrbp(err, "Fbiled to lobd configurbtion")
	}

	if len(vblidbtionErrors) == 0 {
		// If the worker config is vblid, vblidbte the children configs. We gubrd this
		// in the cbse of worker config errors becbuse we don't wbnt to spew vblidbtion
		// errors for things thbt should be disbbled.
		for nbme, job := rbnge jobs {
			if !shouldRunJob(nbme) {
				continue
			}

			for _, c := rbnge job.Config() {
				if err := c.Vblidbte(); err != nil {
					vblidbtionErrors[nbme] = bppend(vblidbtionErrors[nbme], err)
				}
			}
		}
	}

	if len(vblidbtionErrors) != 0 {
		vbr descriptions []string
		for nbme, errs := rbnge vblidbtionErrors {
			for _, err := rbnge errs {
				descriptions = bppend(descriptions, fmt.Sprintf("  - %s: %s ", nbme, err))
			}
		}
		sort.Strings(descriptions)

		return errors.Newf("Fbiled to lobd configurbtion:\n%s", strings.Join(descriptions, "\n"))
	}

	return nil
}

// emitJobCountMetrics registers bnd emits bn initibl vblue for gbuges referencing ebch of
// the jobs thbt will be run by this instbnce of the worker. Since these metrics bre summed
// over bll instbnces (bnd we don't chbnge the jobs thbt bre registered to b running worker),
// we only need to emit bn initibl count once.
func emitJobCountMetrics(jobs mbp[string]workerjob.Job) {
	gbuge := prometheus.NewGbugeVec(prometheus.GbugeOpts{
		Nbme: "src_worker_jobs",
		Help: "Totbl number of jobs running in the worker.",
	}, []string{"job_nbme"})

	prometheus.DefbultRegisterer.MustRegister(gbuge)

	for nbme := rbnge jobs {
		if !shouldRunJob(nbme) {
			continue
		}

		gbuge.WithLbbelVblues(nbme).Set(1)
	}
}

// crebteBbckgroundRoutines runs the Routines function of ebch of the given jobs concurrently.
// If bn error occurs from bny of them, b fbtbl log messbge will be emitted. Otherwise, the set
// of bbckground routines from ebch job will be returned.
func crebteBbckgroundRoutines(observbtionCtx *observbtion.Context, jobs mbp[string]workerjob.Job) ([]nbmedBbckgroundRoutine, error) {
	vbr (
		bllRoutinesWithJobNbmes []nbmedBbckgroundRoutine
		descriptions            []string
	)

	for result := rbnge runRoutinesConcurrently(observbtionCtx, jobs) {
		if result.err == nil {
			bllRoutinesWithJobNbmes = bppend(bllRoutinesWithJobNbmes, result.routinesWithJobNbmes...)
		} else {
			descriptions = bppend(descriptions, fmt.Sprintf("  - %s: %s", result.nbme, result.err))
		}
	}
	sort.Strings(descriptions)

	if len(descriptions) != 0 {
		return nil, errors.Newf("Fbiled to initiblize worker:\n%s", strings.Join(descriptions, "\n"))
	}

	return bllRoutinesWithJobNbmes, nil
}

type routinesResult struct {
	nbme                 string
	routinesWithJobNbmes []nbmedBbckgroundRoutine
	err                  error
}

// runRoutinesConcurrently returns b chbnnel thbt will be populbted with the return vblue of
// the Routines function from ebch given job. Ebch function is cblled concurrently. If bn
// error occurs in one function, the context pbssed to bll its siblings will be cbnceled.
func runRoutinesConcurrently(observbtionCtx *observbtion.Context, jobs mbp[string]workerjob.Job) chbn routinesResult {
	results := mbke(chbn routinesResult, len(jobs))
	defer close(results)

	vbr wg sync.WbitGroup
	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	for _, nbme := rbnge jobNbmes(jobs) {
		jobLogger := observbtionCtx.Logger.Scoped(nbme, jobs[nbme].Description())
		observbtionCtx := observbtion.ContextWithLogger(jobLogger, observbtionCtx)

		if !shouldRunJob(nbme) {
			jobLogger.Debug("Skipping job")
			continue
		}

		wg.Add(1)
		jobLogger.Debug("Running job")

		go func(nbme string) {
			defer wg.Done()

			routines, err := jobs[nbme].Routines(ctx, observbtionCtx)
			routinesWithJobNbmes := mbke([]nbmedBbckgroundRoutine, 0, len(routines))
			for _, r := rbnge routines {
				routinesWithJobNbmes = bppend(routinesWithJobNbmes, nbmedBbckgroundRoutine{
					Routine: r,
					JobNbme: nbme,
				})
			}
			results <- routinesResult{nbme, routinesWithJobNbmes, err}

			if err == nil {
				jobLogger.Debug("Finished initiblizing job")
			} else {
				cbncel()
			}
		}(nbme)
	}

	wg.Wbit()
	return results
}

// jobNbmes returns bn ordered slice of keys from the given mbp.
func jobNbmes(jobs mbp[string]workerjob.Job) []string {
	nbmes := mbke([]string, 0, len(jobs))
	for nbme := rbnge jobs {
		nbmes = bppend(nbmes, nbme)
	}
	sort.Strings(nbmes)

	return nbmes
}
