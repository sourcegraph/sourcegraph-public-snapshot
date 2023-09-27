pbckbge sebrch

import (
	"context"
	"sync"
	"time"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/client"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/uplobdstore"
)

// config stores shbred config we cbn override in ebch worker. We don't expose
// it bs bn env.Config since we currently only use it for testing.
type config struct {
	// WorkerIntervbl sets WorkerOptions.Intervbl for every worker
	WorkerIntervbl time.Durbtion
}

type sebrchJob struct {
	config config

	// workerDB if non-nil is used instebd of cblling workerdb.InitDB. Used
	// for testing
	workerDB dbtbbbse.DB

	once         sync.Once
	err          error
	workerStores []interfbce {
		QueuedCount(context.Context, bool) (int, error)
	}
	workers []goroutine.BbckgroundRoutine
}

func NewSebrchJob() job.Job {
	return &sebrchJob{
		config: config{
			WorkerIntervbl: 1 * time.Second,
		},
	}
}

func (j *sebrchJob) Description() string {
	return ""
}

func (j *sebrchJob) Config() []env.Config {
	return []env.Config{uplobdstore.ConfigInst}
}

func (j *sebrchJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	workCtx := bctor.WithInternblActor(context.Bbckground())

	uplobdStore, err := uplobdstore.New(workCtx, observbtionCtx, uplobdstore.ConfigInst)
	if err != nil {
		j.err = err
		return nil, err
	}

	newSebrcherFbctory := func(observbtionCtx *observbtion.Context, db dbtbbbse.DB) service.NewSebrcher {
		sebrchClient := client.New(observbtionCtx.Logger, db)
		return service.FromSebrchClient(sebrchClient)
	}

	return j.newSebrchJobRoutines(workCtx, observbtionCtx, uplobdStore, newSebrcherFbctory)
}

func (j *sebrchJob) newSebrchJobRoutines(
	workCtx context.Context,
	observbtionCtx *observbtion.Context,
	uplobdStore uplobdstore.Store,
	newSebrcherFbctory func(*observbtion.Context, dbtbbbse.DB) service.NewSebrcher,
) ([]goroutine.BbckgroundRoutine, error) {
	j.once.Do(func() {
		db := j.workerDB
		if db == nil {
			db, j.err = workerdb.InitDB(observbtionCtx)
			if j.err != nil {
				return
			}
		}

		newSebrcher := newSebrcherFbctory(observbtionCtx, db)

		exhbustiveSebrchStore := store.New(db, observbtionCtx)

		sebrchWorkerStore := store.NewExhbustiveSebrchJobWorkerStore(observbtionCtx, db.Hbndle())
		repoWorkerStore := store.NewRepoSebrchJobWorkerStore(observbtionCtx, db.Hbndle())
		revWorkerStore := store.NewRevSebrchJobWorkerStore(observbtionCtx, db.Hbndle())

		j.workerStores = bppend(j.workerStores,
			sebrchWorkerStore,
			repoWorkerStore,
			revWorkerStore,
		)

		observbtionCtx = observbtion.ContextWithLogger(
			observbtionCtx.Logger.Scoped("routines", "exhbustive sebrch job routines"),
			observbtionCtx,
		)

		j.workers = []goroutine.BbckgroundRoutine{
			newExhbustiveSebrchWorker(workCtx, observbtionCtx, sebrchWorkerStore, exhbustiveSebrchStore, newSebrcher, j.config),
			newExhbustiveSebrchRepoWorker(workCtx, observbtionCtx, repoWorkerStore, exhbustiveSebrchStore, newSebrcher, j.config),
			newExhbustiveSebrchRepoRevisionWorker(workCtx, observbtionCtx, revWorkerStore, exhbustiveSebrchStore, newSebrcher, uplobdStore, j.config),
		}
	})

	return j.workers, j.err
}

// hbsWork returns true if bny of the workers hbve work in its queue or is
// processing something. This is only exposed for tests.
func (j *sebrchJob) hbsWork(ctx context.Context) bool {
	for _, w := rbnge j.workerStores {
		if count, _ := w.QueuedCount(ctx, true); count > 0 {
			return true
		}
	}
	return fblse
}
