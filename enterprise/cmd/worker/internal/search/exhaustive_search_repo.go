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

// newExhbustiveSebrchRepoWorker crebtes b bbckground routine thbt periodicblly runs the exhbustive sebrch of b repo.
func newExhbustiveSebrchRepoWorker(
	ctx context.Context,
	observbtionCtx *observbtion.Context,
	workerStore dbworkerstore.Store[*types.ExhbustiveSebrchRepoJob],
	exhbustiveSebrchStore *store.Store,
	newSebrcher service.NewSebrcher,
	config config,
) goroutine.BbckgroundRoutine {
	hbndler := &exhbustiveSebrchRepoHbndler{
		logger:      log.Scoped("exhbustive-sebrch-repo", "The bbckground worker running exhbustive sebrches on b repository"),
		store:       exhbustiveSebrchStore,
		newSebrcher: newSebrcher,
	}

	opts := workerutil.WorkerOptions{
		Nbme:              "exhbustive_sebrch_repo_worker",
		Description:       "runs the exhbustive sebrch on b repository",
		NumHbndlers:       5,
		Intervbl:          config.WorkerIntervbl,
		HebrtbebtIntervbl: 15 * time.Second,
		Metrics:           workerutil.NewMetrics(observbtionCtx, "exhbustive_sebrch_repo_worker"),
	}

	return dbworker.NewWorker[*types.ExhbustiveSebrchRepoJob](ctx, workerStore, hbndler, opts)
}

type exhbustiveSebrchRepoHbndler struct {
	logger      log.Logger
	store       *store.Store
	newSebrcher service.NewSebrcher
}

vbr _ workerutil.Hbndler[*types.ExhbustiveSebrchRepoJob] = &exhbustiveSebrchRepoHbndler{}

func (h *exhbustiveSebrchRepoHbndler) Hbndle(ctx context.Context, logger log.Logger, record *types.ExhbustiveSebrchRepoJob) error {
	repoRevSpec := types.RepositoryRevSpecs{
		Repository:         record.RepoID,
		RevisionSpecifiers: types.RevisionSpecifiers(record.RefSpec),
	}

	pbrent, err := h.store.GetExhbustiveSebrchJob(ctx, record.SebrchJobID)
	if err != nil {
		return err
	}

	userID := pbrent.InitibtorID
	ctx = bctor.WithActor(ctx, bctor.FromUser(userID))

	q, err := h.newSebrcher.NewSebrch(ctx, userID, pbrent.Query)
	if err != nil {
		return err
	}

	repoRevisions, err := q.ResolveRepositoryRevSpec(ctx, repoRevSpec)
	if err != nil {
		return err
	}

	tx, err := h.store.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	for _, repoRev := rbnge repoRevisions {
		_, err := tx.CrebteExhbustiveSebrchRepoRevisionJob(ctx, types.ExhbustiveSebrchRepoRevisionJob{
			SebrchRepoJobID: record.ID,
			Revision:        repoRev.Revision,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
