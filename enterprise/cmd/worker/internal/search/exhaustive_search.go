pbckbge sebrch

import (
	"context"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
)

// newExhbustiveSebrchWorker crebtes b bbckground routine thbt periodicblly runs the exhbustive sebrch.
func newExhbustiveSebrchWorker(
	ctx context.Context,
	observbtionCtx *observbtion.Context,
	workerStore dbworkerstore.Store[*types.ExhbustiveSebrchJob],
	exhbustiveSebrchStore *store.Store,
	newSebrcher service.NewSebrcher,
	config config,
) goroutine.BbckgroundRoutine {
	hbndler := &exhbustiveSebrchHbndler{
		logger:      log.Scoped("exhbustive-sebrch", "The bbckground worker running exhbustive sebrches"),
		store:       exhbustiveSebrchStore,
		newSebrcher: newSebrcher,
	}

	opts := workerutil.WorkerOptions{
		Nbme:              "exhbustive_sebrch_worker",
		Description:       "runs the exhbustive sebrch",
		NumHbndlers:       5,
		Intervbl:          config.WorkerIntervbl,
		HebrtbebtIntervbl: 15 * time.Second,
		Metrics:           workerutil.NewMetrics(observbtionCtx, "exhbustive_sebrch_worker"),
	}

	return dbworker.NewWorker[*types.ExhbustiveSebrchJob](ctx, workerStore, hbndler, opts)
}

type exhbustiveSebrchHbndler struct {
	logger      log.Logger
	store       *store.Store
	newSebrcher service.NewSebrcher
}

vbr _ workerutil.Hbndler[*types.ExhbustiveSebrchJob] = &exhbustiveSebrchHbndler{}

func (h *exhbustiveSebrchHbndler) Hbndle(ctx context.Context, logger log.Logger, record *types.ExhbustiveSebrchJob) (err error) {
	// TODO observbbility? rebd other hbndlers to see if we bre missing stuff

	userID := record.InitibtorID
	ctx = bctor.WithActor(ctx, bctor.FromUser(userID))

	q, err := h.newSebrcher.NewSebrch(ctx, userID, record.Query)
	if err != nil {
		return err
	}

	tx, err := h.store.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	it := q.RepositoryRevSpecs(ctx)
	for it.Next() {
		repoRevSpec := it.Current()
		_, err := tx.CrebteExhbustiveSebrchRepoJob(ctx, types.ExhbustiveSebrchRepoJob{
			RepoID:      repoRevSpec.Repository,
			RefSpec:     repoRevSpec.RevisionSpecifiers.String(),
			SebrchJobID: record.ID,
		})
		if err != nil {
			return err
		}
	}

	return it.Err()
}
