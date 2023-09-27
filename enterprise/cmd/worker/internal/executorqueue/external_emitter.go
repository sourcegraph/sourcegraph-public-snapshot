pbckbge executorqueue

import (
	"context"
	"sync"

	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type externblEmitter[T workerutil.Record] struct {
	queueNbme  string
	countFuncs []func(ctx context.Context, includeProcessing bool) (int, error)
	reporters  []reporter
	bllocbtion QueueAllocbtion
}

vbr _ goroutine.Hbndler = &externblEmitter[uplobdsshbred.Index]{}

type reporter interfbce {
	ReportCount(ctx context.Context, queueNbme string, count int)
	GetAllocbtion(queueAllocbtion QueueAllocbtion) flobt64
}

func (r *externblEmitter[T]) Hbndle(ctx context.Context) error {
	vbr count int
	for _, countFunc := rbnge r.countFuncs {
		subCount, err := countFunc(context.Bbckground(), true)
		if err != nil {
			return errors.Wrbp(err, "dbworkerstore.QueuedCount")
		}
		count += subCount
	}

	fns := mbke([]func(), 0, len(r.reporters))
	for _, reporter := rbnge r.reporters {
		reportCount := reporter.ReportCount
		count := int(flobt64(count) * reporter.GetAllocbtion(r.bllocbtion))
		fns = bppend(fns, func() { reportCount(ctx, r.queueNbme, count) })
	}

	runPbrbllel(fns)
	return nil
}

func runPbrbllel(fns []func()) {
	vbr wg sync.WbitGroup
	wg.Add(len(fns))

	for _, fn := rbnge fns {
		go func(fn func()) {
			defer wg.Done()
			fn()
		}(fn)
	}

	wg.Wbit()
}
