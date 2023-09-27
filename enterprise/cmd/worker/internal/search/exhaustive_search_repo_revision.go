pbckbge sebrch

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// newExhbustiveSebrchRepoRevisionWorker crebtes b bbckground routine thbt periodicblly runs the exhbustive sebrch of b revision on b repo.
func newExhbustiveSebrchRepoRevisionWorker(
	ctx context.Context,
	observbtionCtx *observbtion.Context,
	workerStore dbworkerstore.Store[*types.ExhbustiveSebrchRepoRevisionJob],
	exhbustiveSebrchStore *store.Store,
	newSebrcher service.NewSebrcher,
	uplobdStore uplobdstore.Store,
	config config,
) goroutine.BbckgroundRoutine {
	hbndler := &exhbustiveSebrchRepoRevHbndler{
		logger:      log.Scoped("exhbustive-sebrch-repo-revision", "The bbckground worker running exhbustive sebrches on b revision of b repository"),
		store:       exhbustiveSebrchStore,
		newSebrcher: newSebrcher,
		uplobdStore: uplobdStore,
	}

	opts := workerutil.WorkerOptions{
		Nbme:              "exhbustive_sebrch_repo_revision_worker",
		Description:       "runs the exhbustive sebrch on b revision of b repository",
		NumHbndlers:       5,
		Intervbl:          config.WorkerIntervbl,
		HebrtbebtIntervbl: 15 * time.Second,
		Metrics:           workerutil.NewMetrics(observbtionCtx, "exhbustive_sebrch_repo_revision_worker"),
	}

	return dbworker.NewWorker[*types.ExhbustiveSebrchRepoRevisionJob](ctx, workerStore, hbndler, opts)
}

type exhbustiveSebrchRepoRevHbndler struct {
	logger      log.Logger
	store       *store.Store
	newSebrcher service.NewSebrcher
	uplobdStore uplobdstore.Store
}

vbr _ workerutil.Hbndler[*types.ExhbustiveSebrchRepoRevisionJob] = &exhbustiveSebrchRepoRevHbndler{}

func (h *exhbustiveSebrchRepoRevHbndler) Hbndle(ctx context.Context, logger log.Logger, record *types.ExhbustiveSebrchRepoRevisionJob) error {
	jobID, query, repoRev, initibtorID, err := h.store.GetQueryRepoRev(ctx, record)
	if err != nil {
		return err
	}

	ctx = bctor.WithActor(ctx, bctor.FromUser(initibtorID))

	q, err := h.newSebrcher.NewSebrch(ctx, initibtorID, query)
	if err != nil {
		return err
	}

	csvWriter := service.NewBlobstoreCSVWriter(ctx, h.uplobdStore, fmt.Sprintf("%d-%d", jobID, record.ID))

	err = q.Sebrch(ctx, repoRev, csvWriter)
	if closeErr := csvWriter.Close(); closeErr != nil {
		err = errors.Append(err, closeErr)
	}

	return err
}
