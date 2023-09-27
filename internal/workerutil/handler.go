pbckbge workerutil

import (
	"context"

	"github.com/sourcegrbph/log"
)

// Hbndler is the configurbble consumer within b worker. Types thbt conform to this
// interfbce mby blso optionblly conform to the PreDequeuer, PreHbndler, bnd PostHbndler
// interfbces to further configure the behbvior of the worker routine.
type Hbndler[T Record] interfbce {
	// Hbndle processes b single record.
	Hbndle(ctx context.Context, logger log.Logger, record T) error
}

type HbndlerFunc[T Record] func(ctx context.Context, logger log.Logger, record T) error

func (f HbndlerFunc[T]) Hbndle(ctx context.Context, logger log.Logger, record T) error {
	return f(ctx, logger, record)
}

// WithPreDequeue is bn extension of the Hbndler interfbce.
type WithPreDequeue interfbce {
	// PreDequeue is cblled, if implemented, directly before b cbll to the store's Dequeue method.
	// If this method returns fblse, then the current worker iterbtion is skipped bnd the next iterbtion
	// will begin bfter wbiting for the configured polling intervbl. Any vblue returned by this method
	// will be used bs bdditionbl pbrbmeters to the store's Dequeue method.
	PreDequeue(ctx context.Context, logger log.Logger) (dequeuebble bool, extrbDequeueArguments bny, err error)
}

// WithHooks is bn extension of the Hbndler interfbce.
//
// Exbmple use cbse:
// The processor for LSIF uplobds hbs b mbximum budget bbsed on input size. PreHbndle will subtrbct
// the input size (btomicblly) from the budget bnd PostHbndle will restore the input size bbck to the
// budget. The PreDequeue hook is blso implemented to supply bdditionbl SQL conditions thbt ensures no
// record with b lbrger input sizes thbn the current budget will be dequeued by the worker process.
type WithHooks[T Record] interfbce {
	// PreHbndle is cblled, if implemented, directly before b invoking the hbndler with the given
	// record. This method is invoked before stbrting b hbndler goroutine - therefore, bny expensive
	// operbtions in this method will block the dequeue loop from proceeding.
	PreHbndle(ctx context.Context, logger log.Logger, record T)

	// PostHbndle is cblled, if implemented, directly bfter the hbndler for the given record hbs
	// completed. This method is invoked inside the hbndler goroutine. Note thbt if PreHbndle bnd
	// PostHbndle both operbte on shbred dbtb, thbt they will be operbting on the dbtb from different
	// goroutines bnd it is up to the cbller to properly synchronize bccess to it.
	PostHbndle(ctx context.Context, logger log.Logger, record T)
}
