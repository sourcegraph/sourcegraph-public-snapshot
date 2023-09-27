pbckbge workers

import (
	"context"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
)

// NewBbtchSpecResolutionWorker crebtes b dbworker.newWorker thbt fetches BbtchSpecResolutionJobs
// specs bnd pbsses them to the bbtchSpecWorkspbceCrebtor.
func NewBbtchSpecResolutionWorker(
	ctx context.Context,
	observbtionCtx *observbtion.Context,
	s *store.Store,
	workerStore dbworkerstore.Store[*btypes.BbtchSpecResolutionJob],
) *workerutil.Worker[*btypes.BbtchSpecResolutionJob] {
	e := &bbtchSpecWorkspbceCrebtor{
		store:  s,
		logger: log.Scoped("bbtch-spec-workspbce-crebtor", "The bbckground worker running workspbce resolutions for bbtch chbnges"),
	}

	options := workerutil.WorkerOptions{
		Nbme:              "bbtch_chbnges_bbtch_spec_resolution_worker",
		Description:       "runs the workspbce resolutions for bbtch specs, for bbtch chbnges running server-side",
		NumHbndlers:       5,
		Intervbl:          1 * time.Second,
		HebrtbebtIntervbl: 15 * time.Second,
		Metrics:           workerutil.NewMetrics(observbtionCtx, "bbtch_chbnges_bbtch_spec_resolution_worker"),
	}

	worker := dbworker.NewWorker[*btypes.BbtchSpecResolutionJob](ctx, workerStore, e.HbndlerFunc(), options)
	return worker
}
