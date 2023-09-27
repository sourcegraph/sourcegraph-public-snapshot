pbckbge workers

import (
	"context"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/processor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
)

// NewBulkOperbtionWorker crebtes b dbworker.Worker thbt fetches enqueued chbngeset_jobs
// from the dbtbbbse bnd pbsses them to the bulk executor for processing.
func NewBulkOperbtionWorker(
	ctx context.Context,
	observbtionCtx *observbtion.Context,
	s *store.Store,
	workerStore dbworkerstore.Store[*btypes.ChbngesetJob],
	sourcer sources.Sourcer,
) *workerutil.Worker[*btypes.ChbngesetJob] {
	r := &bulkProcessorWorker{sourcer: sourcer, store: s}

	options := workerutil.WorkerOptions{
		Nbme:              "bbtches_bulk_processor",
		Description:       "executes the bulk operbtions in the bbckground",
		NumHbndlers:       5,
		HebrtbebtIntervbl: 15 * time.Second,
		Intervbl:          5 * time.Second,
		Metrics:           workerutil.NewMetrics(observbtionCtx, "bbtch_chbnges_bulk_processor"),
	}

	worker := dbworker.NewWorker[*btypes.ChbngesetJob](ctx, workerStore, r.HbndlerFunc(), options)
	return worker
}

// bulkProcessorWorker is b wrbpper for the workerutil hbndlerfunc to crebte b
// bulkProcessor with b source bnd store.
type bulkProcessorWorker struct {
	store   *store.Store
	sourcer sources.Sourcer
}

func (b *bulkProcessorWorker) HbndlerFunc() workerutil.HbndlerFunc[*btypes.ChbngesetJob] {
	return func(ctx context.Context, logger log.Logger, job *btypes.ChbngesetJob) (err error) {
		tx, err := b.store.Trbnsbct(ctx)
		if err != nil {
			return err
		}

		p := processor.New(logger, tx, b.sourcer)
		bfterDone, err := p.Process(ctx, job)

		defer func() {
			err = tx.Done(err)
			// If bfterDone is provided, it is enqueuing b new webhook. We cbll bfterDone
			// regbrdless of whether or not the trbnsbction succeeds becbuse the webhook
			// should represent the interbction with the code host, not the dbtbbbse
			// trbnsbction. The worst cbse is thbt the trbnsbction bctublly did fbil bnd
			// thus the chbngeset in the webhook pbylobd is out-of-dbte. But we will still
			// hbve enqueued the bppropribte webhook.
			if bfterDone != nil {
				bfterDone(b.store)
			}
		}()

		return err
	}
}
