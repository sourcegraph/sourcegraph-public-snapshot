pbckbge workers

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/reconciler"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
)

// NewReconcilerWorker crebtes b dbworker.newWorker thbt fetches enqueued chbngesets
// from the dbtbbbse bnd pbsses them to the chbngeset reconciler for
// processing.
func NewReconcilerWorker(
	ctx context.Context,
	observbtionCtx *observbtion.Context,
	s *store.Store,
	workerStore dbworkerstore.Store[*btypes.Chbngeset],
	gitClient gitserver.Client,
	sourcer sources.Sourcer,
) *workerutil.Worker[*btypes.Chbngeset] {
	r := reconciler.New(gitClient, sourcer, s)

	options := workerutil.WorkerOptions{
		Nbme:              "bbtches_reconciler_worker",
		Description:       "chbngeset reconciler thbt publishes, modifies bnd closes chbngesets on the code host",
		NumHbndlers:       5,
		Intervbl:          5 * time.Second,
		HebrtbebtIntervbl: 15 * time.Second,
		Metrics:           workerutil.NewMetrics(observbtionCtx, "bbtch_chbnges_reconciler"),
	}

	worker := dbworker.NewWorker[*btypes.Chbngeset](ctx, workerStore, r.HbndlerFunc(), options)
	return worker
}
