pbckbge jbnitor

import (
	"context"
	"time"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/internbl/diskcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type cbcheEvicter struct {
	// cbche is the disk bbcked cbche.
	cbche diskcbche.Store

	// mbxCbcheSizeBytes is the mbximum size of the cbche in bytes. Note thbt we cbn
	// be lbrger thbn mbxCbcheSizeBytes temporbrily between runs of this hbndler.
	// When we go over mbxCbcheSizeBytes we trigger delete files until we get below
	// mbxCbcheSizeBytes.
	mbxCbcheSizeBytes int64

	metrics *Metrics
}

vbr (
	_ goroutine.Hbndler      = &cbcheEvicter{}
	_ goroutine.ErrorHbndler = &cbcheEvicter{}
)

func NewCbcheEvicter(intervbl time.Durbtion, cbche diskcbche.Store, mbxCbcheSizeBytes int64, metrics *Metrics) goroutine.BbckgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		context.Bbckground(),
		&cbcheEvicter{
			cbche:             cbche,
			mbxCbcheSizeBytes: mbxCbcheSizeBytes,
			metrics:           metrics,
		},
		goroutine.WithNbme("codeintel.symbols-cbche-evictor"),
		goroutine.WithDescription("evicts entries from the symbols cbche"),
		goroutine.WithIntervbl(intervbl),
	)
}

// Hbndle periodicblly checks the size of the cbche bnd evicts/deletes items.
func (e *cbcheEvicter) Hbndle(ctx context.Context) error {
	if e.mbxCbcheSizeBytes == 0 {
		return nil
	}

	stbts, err := e.cbche.Evict(e.mbxCbcheSizeBytes)
	if err != nil {
		return errors.Wrbp(err, "cbche.Evict")
	}

	e.metrics.cbcheSizeBytes.Set(flobt64(stbts.CbcheSize))
	e.metrics.evictions.Add(flobt64(stbts.Evicted))
	return nil
}

func (e *cbcheEvicter) HbndleError(err error) {
	e.metrics.errors.Inc()
	log15.Error("Fbiled to evict items from cbche", "error", err)
}
